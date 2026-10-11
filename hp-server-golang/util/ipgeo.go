package util

import (
	"errors"
	"io/fs"
	"net"
	"strconv"
	"strings"
	"sync"

	"hp-server-lib/log"

	"github.com/oschwald/geoip2-golang"
)

// ipGeoNameLang MaxMind 库里的国家 / 城市名是多语言的，这里取展示名的优先级。
// 大屏是中文界面，优先 zh-CN；少数条目没有中文名，逐级兜底。
var ipGeoNameLang = []string{"zh-CN", "zh", "en"}

// IpRegion IP 归属地（GeoLite2 / GeoIP2 库，替代原 ip2region）。
//
// 相比旧的 ip2region 实现：
//   - 多了经纬度 —— 前端可以直接拿来在地图上打点；
//   - 少了运营商 —— GeoLite2-City 不含 ISP 信息，需要额外加载 GeoLite2-ASN.mmdb 才有。
//
// 一个 GeoLite2-City 库同时覆盖 IPv4 和 IPv6，不再像 ip2region 那样分两个文件。
type IpRegion struct {
	Country     string  // 国家，如：中国
	Province    string  // 一级行政区：省 / 州
	City        string  // 城市
	Isp         string  // 运营商 / 自治域组织（未加载 ASN 库时为空）
	CountryCode string  // 两位国家代码，如 CN / US（前端可据此显示国旗）
	Latitude    float64 // 纬度，0 表示未定位（含城市级查不到的情况）
	Longitude   float64 // 经度，0 表示未定位
	Raw         string  // 原始串，便于排查
}

var (
	// ipGeoMu 保护下面三个句柄，支持运行中热替换（重新加载归属地库时替换指针）
	ipGeoMu sync.RWMutex
	// ipGeoCity GeoLite2-City：国家 / 省 / 城市 / 经纬度
	ipGeoCity *geoip2.Reader
	// ipGeoAsn GeoLite2-ASN：运营商。可选，没配套 *.mmdb 时为 nil
	ipGeoAsn   *geoip2.Reader
	ipGeoReady bool

	// ipGeoCache IP -> 归属地，避免同一个 IP 反复走树查找。
	// 查不到的结果也会缓存，防止漏掉的 IP 每次请求都白查一遍库
	ipGeoCacheMu sync.RWMutex
	ipGeoCache   = map[string]IpRegion{}
	// ipGeoCacheMax 缓存上限，超过后整体清空，防止高并发下无限膨胀
	ipGeoCacheMax = 100000
)

// InitIpGeoFromFS 从文件系统加载 MaxMind 的 mmdb 库。
// 声明成 fs.FS 而不是 embed.FS：方便测试和改成从磁盘读取，
// 实际调用处传的是随程序一起打包的内嵌文件系统（无需外部文件）。
// cityPath 是 GeoLite2-City 库的路径，必填 —— 国家 / 省 / 城市 / 经纬度全靠它；
// asnPath 是 GeoLite2-ASN 库的路径，可选 —— 只有需要运营商信息时才填，留空或未打包就跳过。
// 一个 City 库同时覆盖 IPv4 和 IPv6；加载失败只记录日志不影响启动，归属地字段会留空。
func InitIpGeoFromFS(fsys fs.FS, cityPath, asnPath string) error {
	ipGeoMu.Lock()
	defer ipGeoMu.Unlock()

	ipGeoCity = nil
	ipGeoAsn = nil
	ipGeoReady = false
	// 库换了，旧缓存必须一并失效，否则会继续吐出上一个版本的归属地
	ipGeoCacheMu.Lock()
	ipGeoCache = map[string]IpRegion{}
	ipGeoCacheMu.Unlock()

	city, err := openIpGeoDB(fsys, cityPath)
	if err != nil {
		return err
	}
	ipGeoCity = city
	ipGeoReady = true
	log.Infof("IP 归属地库加载成功：%s|类型:%s", cityPath, city.Metadata().DatabaseType)

	// ASN 库缺失属于正常情况（不少部署只配 City），不打错误日志
	if strings.TrimSpace(asnPath) != "" {
		asn, err := openIpGeoDB(fsys, asnPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				log.Errorf("ASN 库加载失败，运营商字段将为空：%v", err)
			}
		} else {
			ipGeoAsn = asn
			log.Infof("ASN 库加载成功：%s|类型:%s", asnPath, asn.Metadata().DatabaseType)
		}
	}

	return nil
}

// IpGeoReady 归属地库是否已就绪
func IpGeoReady() bool {
	ipGeoMu.RLock()
	defer ipGeoMu.RUnlock()
	return ipGeoReady
}

// LookupIpRegion 查询 IP 归属地。未加载库 / 非法 IP / 库里查不到时返回 false。
//
// geoip2 的 Reader 本身并发安全（每次查询分配的都是新结构体），不需要像
// 旧的 xdb.Searcher 那样用池隔离。
func LookupIpRegion(ip string) (IpRegion, bool) {
	if ip == "" {
		return IpRegion{}, false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return IpRegion{}, false
	}

	// 本机与内网不查库，直接给出可读结果
	if parsed.IsLoopback() {
		return IpRegion{Country: "本机", Province: "本机", City: "本机", Isp: "本机", Raw: "本机"}, true
	}
	if parsed.IsPrivate() {
		return IpRegion{Country: "局域网", Province: "局域网", City: "局域网", Isp: "局域网", Raw: "局域网"}, true
	}

	if region, ok := loadIpGeoCache(ip); ok {
		return region, region.Country != ""
	}

	ipGeoMu.RLock()
	cityDB := ipGeoCity
	asnDB := ipGeoAsn
	ipGeoMu.RUnlock()
	if cityDB == nil {
		return IpRegion{}, false
	}

	region, err := buildIpRegion(cityDB, asnDB, parsed)
	if err != nil {
		// 库里没有这条记录：缓存空结果，同一个 IP 不会每来一次都白跑一遍树查找
		storeIpGeoCache(ip, IpRegion{})
		return IpRegion{}, false
	}

	storeIpGeoCache(ip, region)
	return region, region.Country != ""
}

// openIpGeoDB 从内嵌文件系统读出一个 mmdb 库。
// FromBytes 直接把整块字节当查找数据用（不复制），所以内存占用约等于文件大小。
func openIpGeoDB(fsys fs.FS, path string) (*geoip2.Reader, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("mmdb 路径为空")
	}
	// 用 fs.ReadFile 而不是 fsys.ReadFile：前者对任意 fs.FS 都成立，便于测试和换成磁盘目录
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	db, err := geoip2.FromBytes(content)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// buildIpRegion 把 City（可选叠加 ASN）的查询结果摊平成 IpRegion
func buildIpRegion(cityDB, asnDB *geoip2.Reader, ip net.IP) (IpRegion, error) {
	city, err := cityDB.City(ip)
	if err != nil {
		return IpRegion{}, err
	}

	region := IpRegion{
		Country:     pickIpGeoName(city.Country.Names),
		CountryCode: city.Country.IsoCode,
		Latitude:    city.Location.Latitude,
		Longitude:   city.Location.Longitude,
	}
	// 部分 IP（如军事 / 卫星网络）只有注册国，没有实际归属地，退回注册国
	if region.Country == "" {
		region.Country = pickIpGeoName(city.RegisteredCountry.Names)
		region.CountryCode = city.RegisteredCountry.IsoCode
	}
	// subdivisions 是从大到小排列的行政区，取第一个即省 / 州一级
	if len(city.Subdivisions) > 0 {
		region.Province = pickIpGeoName(city.Subdivisions[0].Names)
	}
	region.City = pickIpGeoName(city.City.Names)

	if asnDB != nil {
		if asn, err := asnDB.ASN(ip); err == nil && asn != nil {
			if asn.AutonomousSystemOrganization != "" {
				region.Isp = asn.AutonomousSystemOrganization
			} else if asn.AutonomousSystemNumber > 0 {
				// 少数自治域没登记组织名，退回 AS 号
				region.Isp = "AS" + strconv.FormatUint(uint64(asn.AutonomousSystemNumber), 10)
			}
		}
	}

	region.Raw = buildIpRegionRaw(region)
	return region, nil
}

// buildIpRegionRaw 拼原始串，排查时打印这一条就够了
func buildIpRegionRaw(r IpRegion) string {
	parts := make([]string, 0, 7)
	parts = append(parts, r.Country, r.Province, r.City)
	if r.CountryCode != "" {
		parts = append(parts, r.CountryCode)
	}
	if r.Isp != "" {
		parts = append(parts, r.Isp)
	}
	if r.Latitude != 0 || r.Longitude != 0 {
		parts = append(parts,
			strconv.FormatFloat(r.Latitude, 'f', 4, 64)+","+
				strconv.FormatFloat(r.Longitude, 'f', 4, 64))
	}
	// 过滤空段后再拼，避免连续的分隔符
	nonEmpty := parts[:0]
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, "|")
}

// pickIpGeoName 按语言优先级取 MaxMind 的多语言名
func pickIpGeoName(names map[string]string) string {
	for _, lang := range ipGeoNameLang {
		if v := strings.TrimSpace(names[lang]); v != "" {
			return v
		}
	}
	return ""
}

func loadIpGeoCache(ip string) (IpRegion, bool) {
	ipGeoCacheMu.RLock()
	region, ok := ipGeoCache[ip]
	ipGeoCacheMu.RUnlock()
	return region, ok
}

func storeIpGeoCache(ip string, region IpRegion) {
	ipGeoCacheMu.Lock()
	if len(ipGeoCache) >= ipGeoCacheMax {
		ipGeoCache = make(map[string]IpRegion, 1024)
	}
	ipGeoCache[ip] = region
	ipGeoCacheMu.Unlock()
}
