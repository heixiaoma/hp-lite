package net.hserver.hplite.domian.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import net.hserver.hplite.message.ConnectType;

@Data
@TableName("user_config")
public class UserConfigEntity {

    @TableId(type = IdType.AUTO)
    private Integer id;
    /**
     * 当前key
     */
    private String configKey;

    /**
     * 用户KEY
     */
    private String deviceKey;

    /**
     * 套餐IP
     */
    private String serverIp;

    /**
     * 套餐端口
     */

    private Integer serverPort;

    /**
     * 本地IP
     */
    private String localIp;

    /**
     * 本地端口
     */
    private Integer localPort;

    /**
     * 穿透类型
     */
    private ConnectType connectType;

    /**
     * 备注
     */
    private String remarks;

    /**
     * 端口
     */
    private Integer port;

    /**
     * 域名
     */
    private String domain;

    /**
     * 状态
     */
    private String statusMsg;


    private ProxyVersion proxyVersion=ProxyVersion.NONE;


    /**
     * SSL证书Key
     */
    private String certificateKey;

    /**
     * 证书内容
     */
    private String certificateContent;

    /**
     * 只支持这两种
     */
    public  enum ProxyVersion{
        V1,V2,NONE
    }
}
