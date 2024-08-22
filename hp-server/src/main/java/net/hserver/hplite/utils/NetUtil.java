package net.hserver.hplite.utils;

import java.io.IOException;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.ServerSocket;

/**
 * @author hxm
 */
public class NetUtil {
    /**
     * 获取一个可用的端口
     *
     * @return
     */
    public static int getAvailablePort() {
        ServerSocket serverSocket=null;
        try {
            serverSocket = new ServerSocket(0);
            return serverSocket.getLocalPort();
        } catch (Throwable ignored) {
        }finally {
            if (serverSocket!=null){
                try {
                    serverSocket.close();
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        }
        return -1;
    }

    /**
     * 检查是不是内网地址
     * @param address
     * @return
     */
    public static boolean isPrivateAddress(InetSocketAddress address) {
        InetAddress inetAddress = address.getAddress();
        if (inetAddress != null && !inetAddress.isLinkLocalAddress()) {
            byte[] addrBytes = inetAddress.getAddress();
            int addr = ((addrBytes[0] & 0xFF) << 24) | ((addrBytes[1] & 0xFF) << 16) |
                    ((addrBytes[2] & 0xFF) << 8) | (addrBytes[3] & 0xFF);
            return ((addr >>> 24) == 10) ||
                    (((addr >>> 24) & 0xFF) == 172 && ((addr >>> 16) & 0xF0) == 16) ||
                    (((addr >>> 24) & 0xFF) == 192 && ((addr >>> 16) & 0xFF) == 168)||
                    (addr == 0x7F000001)||
                    (addr == 0x00000000);
        }
        return false;
    }

}
