package net.hserver.hplite.message;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class UserConnectInfo {
    private Integer id;
    private Integer configId;
    //自定义的域名
    private String domain;

    //开通的端口号 -1随机
    private Integer port;


    public UserConnectInfo(Integer id, String email, Integer type,Integer configId) {

        this.configId=configId;
    }
}
