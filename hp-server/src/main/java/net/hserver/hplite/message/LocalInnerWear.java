package net.hserver.hplite.message;

import cn.hutool.crypto.digest.MD5;
import cn.hutool.json.JSONUtil;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class LocalInnerWear {

    /**
     * 穿透服务器IP
     */
    private String serverIp;

    /**
     * 穿透服务器的端口
     */
    private Integer serverPort;

    /**
     * 本地IP
     */
    private String localIp;

    /**
     * 本地的端口
     */
    private Integer localPort;

    /**
     * 本地穿透的key
     */
    private String configKey;

    /**
     * 穿透类型
     */
    private ConnectType connectType;

    private String md5;

    public String getMd5() {
        return MD5.create().digestHex(serverIp + serverPort + localIp + localPort + configKey + connectType.name());
    }

    /**
     * 转换为json
     *
     * @return
     */
    public String toJson() {
        return JSONUtil.toJsonStr(this);
    }
}
