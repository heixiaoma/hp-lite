package net.hserver.hplite;

import cn.hserver.HServerApplication;
import cn.hserver.core.ioc.annotation.HServerBoot;


/**
 * @author hxm
 */
@HServerBoot
public class StartProxy {
    public static void main(String[] args) {
        HServerApplication.run(StartProxy.class, new Integer[]{80, 443, 9090, 6666}, args);
    }
}
