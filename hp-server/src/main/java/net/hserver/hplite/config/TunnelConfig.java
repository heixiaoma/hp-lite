package net.hserver.hplite.config;

import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Value;
import lombok.Data;

@Bean
@Data
public class TunnelConfig {

    @Value("tunnel.ip")
    private String ip;

    @Value("tunnel.port")
    private Integer port;



}
