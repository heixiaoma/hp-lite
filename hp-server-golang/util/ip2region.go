package util

import (
	"embed"
	"errors"
	"io/fs"
	"net"
	"strings"
	"sync"

	"hp-server-lib/log"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// IpRegion IP 归属地。
// ip2region 原始数据格式，其中 0 表示无数据：
//
//	v2 库（Structure20）：国家|区域|省份|城市|ISP
//	v3 库（Structure30）：国家|省份|城市|ISP|国家代码
type IpRegion struct {
	Country     string // 国家，如：中国
	Province    string // 省份 / 州
	City        string // 城市
	Isp         string // 运营商 / 组织
	CountryCode string // 两位国家代码，如 CN / US（仅 v3 库有，前端可据此显示国旗）
	Raw         string // 原始串，便于排查
}

var (
	// ipRegionMu 保护 ipRegionDBs / ipRegionReady
	ipRegionMu sync.RWMutex
	// ipRegionDBs 按 IP 版本存放已加载的库：key 为 xdb.IPv4VersionNo / xdb.IPv6VersionNo
	ipRegionDBs   = map[int]*ipRegionDB{}
	ipRegionReady bool

	// ipRegionCache IP -> 归属地，避免同一个 IP 反复做二分查找
	ipRegionCacheMu sync.RWMutex
	ipRegionCache   = map[string]IpRegion{}
	// ipRegionCacheMax 缓存上限，超过后整体清空，防止高并发下无限膨胀
	ipRegionCacheMax = 100000
)

// ipRegionDB 一个已加载的 xdb 库（整库缓存进内存，查询零 IO）
type ipRegionDB struct {
	version *xdb.Version
	content []byte
	// layout 数据版本（xdb.Structure20 / xdb.Structure30），决定字段解析顺序
	layout int
	// pool 复用 Searcher：xdb.Searcher 并非并发安全（内部有 ioCount 读写），
	// 用池给每次查询分配独立实例，既无锁竞争也无数据竞争
	pool sync.Pool
}

// InitIp2RegionFromFS 从内嵌文件系统加载 ip2region 的 xdb 库（随程序一起打包，无需外部文件）。
// v4Path / v6Path 分别是 IPv4、IPv6 库在 embed.FS 中的路径，留空或未内嵌则跳过；库类型由文件头自动识别。
// 加载失败只记录日志不影响启动，归属地字段会留空。
func InitIp2RegionFromFS(fsys embed.FS, v4Path, v6Path string) error {
	ipRegionMu.Lock()
	defer ipRegionMu.Unlock()

	loaded := make(map[int]*ipRegionDB)
	var failed []string
	for _, path := range []string{v4Path, v6Path} {
		if strings.TrimSpace(path) == "" {
			continue
		}
		content, err := xdb.LoadContentFromFS(fsys, path)
		if err != nil {
			// 没内嵌对应库（例如只打包了 IPv4 库）属于正常情况，不打错误日志
			if errors.Is(err, fs.ErrNotExist) {
				log.Infof("ip2region 未内嵌数据文件，跳过：%s", path)
				continue
			}
			failed = append(failed, path+": "+err.Error())
			continue
		}
		db, err := loadIpRegionDB(content)
		if err != nil {
			failed = append(failed, path+": "+err.Error())
			continue
		}
		loaded[db.version.Id] = db
		log.Infof("ip2region 加载成功：%s|版本:%s", path, db.version.Name)
	}

	ipRegionDBs = loaded
	ipRegionReady = len(loaded) > 0
	if !ipRegionReady {
		return errors.New("ip2region 未加载任何数据文件：" + strings.Join(failed, "；"))
	}
	if len(failed) > 0 {
		log.Errorf("ip2region 部分数据文件加载失败：%s", strings.Join(failed, "；"))
	}
	return nil
}

// Ip2RegionReady 归属地库是否已就绪
func Ip2RegionReady() bool {
	ipRegionMu.RLock()
	defer ipRegionMu.RUnlock()
	return ipRegionReady
}

// LookupIpRegion 查询 IP 归属地。未加载库 / 非法 IP / 库里查不到时返回 false。
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

	if region, ok := loadIpRegionCache(ip); ok {
		return region, region.Country != ""
	}

	isV4 := parsed.To4() != nil
	ipRegionMu.RLock()
	db := ipRegionDBs[ipRegionVersionNo(isV4)]
	ipRegionMu.RUnlock()
	if db == nil {
		return IpRegion{}, false
	}

	// IPv4-mapped 地址（::ffff:1.2.3.4）统一按 IPv4 查询
	query := ip
	if isV4 {
		query = parsed.To4().String()
	}
	raw, err := db.search(query)
	if err != nil {
		return IpRegion{}, false
	}

	region := db.parseIpRegion(raw)
	storeIpRegionCache(ip, region)
	return region, region.Country != ""
}

// loadIpRegionDB 解析整库内容并识别 IP 版本 / 数据版本
func loadIpRegionDB(content []byte) (*ipRegionDB, error) {
	header, err := xdb.LoadHeaderFromBuff(content)
	if err != nil {
		return nil, err
	}
	version, err := xdb.VersionFromHeader(header)
	if err != nil {
		return nil, err
	}
	db := &ipRegionDB{version: version, content: content, layout: int(header.Version)}
	db.pool.New = func() any {
		s, err := xdb.NewWithBuffer(version, content)
		if err != nil {
			return nil
		}
		return s
	}
	return db, nil
}

// search 从池中取一个 Searcher 查询，用完归还
func (db *ipRegionDB) search(ip string) (string, error) {
	s, err := db.acquire()
	if err != nil {
		return "", err
	}
	defer db.pool.Put(s)
	return s.Search(ip)
}

func (db *ipRegionDB) acquire() (*xdb.Searcher, error) {
	if v := db.pool.Get(); v != nil {
		if s, ok := v.(*xdb.Searcher); ok {
			return s, nil
		}
	}
	return xdb.NewWithBuffer(db.version, db.content)
}

func ipRegionVersionNo(isV4 bool) int {
	if isV4 {
		return xdb.IPv4VersionNo
	}
	return xdb.IPv6VersionNo
}

// parseIpRegion 按库的数据版本解析归属地，0 视为无数据
func (db *ipRegionDB) parseIpRegion(raw string) IpRegion {
	region := IpRegion{Raw: raw}
	if raw == "" {
		return region
	}
	parts := strings.Split(raw, "|")
	for i := range parts {
		if parts[i] == "0" {
			parts[i] = ""
		}
	}
	region.Country = at(parts, 0)
	if db.layout >= xdb.Structure30 {
		// v3：国家|省份|城市|ISP|国家代码
		region.Province = at(parts, 1)
		region.City = at(parts, 2)
		region.Isp = at(parts, 3)
		region.CountryCode = at(parts, 4)
		return region
	}
	// v2：国家|区域|省份|城市|ISP，第 2 段"区域"不需要，跳过
	region.Province = at(parts, 2)
	region.City = at(parts, 3)
	region.Isp = at(parts, 4)
	return region
}

func at(parts []string, i int) string {
	if i < 0 || i >= len(parts) {
		return ""
	}
	return parts[i]
}

func loadIpRegionCache(ip string) (IpRegion, bool) {
	ipRegionCacheMu.RLock()
	region, ok := ipRegionCache[ip]
	ipRegionCacheMu.RUnlock()
	return region, ok
}

func storeIpRegionCache(ip string, region IpRegion) {
	ipRegionCacheMu.Lock()
	if len(ipRegionCache) >= ipRegionCacheMax {
		ipRegionCache = make(map[string]IpRegion, 1024)
	}
	ipRegionCache[ip] = region
	ipRegionCacheMu.Unlock()
}
