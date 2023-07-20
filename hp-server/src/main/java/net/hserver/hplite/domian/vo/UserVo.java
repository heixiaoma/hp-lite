package net.hserver.hplite.domian.vo;

import lombok.Data;

/**
 * @author hxm
 */
@Data
public class UserVo {
    //用户ID
    private String id;

    //用户名
    private String username;

    //-1 封号
    private Integer type;

    //官方域名二级名字，前缀
    private String domain;

    //自定义的域名
    private String customDomain;

    //开通的端口号 -1随机
    private Integer port;

}
