package net.hserver.hplite.message;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.ToString;

import java.util.List;

@Data
@AllArgsConstructor
@NoArgsConstructor
@ToString
public class ServerInfo {

    /**
     * 主机地址
     */
    private String host;

    /**
     * 穿透数据端口
     */
    private Integer port;
    /**
     * 服务名
     */
    private String name;
}
