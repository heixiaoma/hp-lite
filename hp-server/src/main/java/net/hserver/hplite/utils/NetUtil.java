package net.hserver.hplite.utils;

import java.io.IOException;
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

}
