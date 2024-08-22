package net.hserver.hplite.config;


import io.netty.util.AttributeKey;

public interface CostConfig {
    AttributeKey<Boolean> flow = AttributeKey.valueOf("flow");

    /**
     * quic协议协商
     */
    String HP_LITE = "HP_LITE";

}
