package net.hserver.hplite;

import cn.hserver.HServerApplication;
import cn.hserver.core.ioc.annotation.HServerBoot;
import cn.hserver.core.server.util.PropUtil;
import cn.hutool.crypto.GlobalBouncyCastleProvider;


/**
 * @author hxm
 */
@HServerBoot
public class StartProxy {
    public static void main(String[] args) {
        GlobalBouncyCastleProvider.setUseBouncyCastle(false);
        PropUtil instance = PropUtil.getInstance();
        Integer adminPort = instance.getInt("admin.port");
        Integer cmdPort = instance.getInt("cmd.port");
        Boolean openDomain = instance.getBoolean("tunnel.openDomain");
        if (openDomain!=null&&openDomain){
            HServerApplication.run(StartProxy.class, new Integer[]{80,443,adminPort, cmdPort}, args);
        }else {
            HServerApplication.run(StartProxy.class, new Integer[]{adminPort, cmdPort}, args);
        }
    }
}
