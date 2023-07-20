package net.hserver.hplite.config;

import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Value;
import lombok.Data;

@Bean
@Data
public class WebConfig {

    @Value("adminAddress")
    private String adminAddress;

    @Value("host")
    private String host;

    @Value("port")
    private Integer port;

    @Value("name")
    private String name;

    @Value("price")
    private Integer price;

    @Value("maxFlowMb")
    private Integer maxFlowMb;

    @Value("modulus")
    private Integer modulus;
}
