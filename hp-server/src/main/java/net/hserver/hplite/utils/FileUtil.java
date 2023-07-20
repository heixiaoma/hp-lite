
package net.hserver.hplite.utils;


import java.io.*;

/**
 * @author hxm
 */
public class FileUtil {

    public static void copyFile(InputStream is, String outPath) {
        try {
            FileOutputStream fos = new FileOutputStream(outPath);
            byte[] b = new byte[1024];
            while ((is.read(b)) != -1) {
                fos.write(b);
            }
            is.close();
            fos.close();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

}
