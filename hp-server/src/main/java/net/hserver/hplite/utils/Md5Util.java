package net.hserver.hplite.utils;

import java.io.IOException;
import java.io.InputStream;
import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

public class Md5Util {

    public static String get(byte[] bytes) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] digest = md.digest(bytes);
            return String.format("%032x", new BigInteger(1, digest));
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException(e);
        }
    }

    public static String get(InputStream inputStream) {
        try {
            int readCount;
            byte[] buffer = new byte[1024];
            MessageDigest md = MessageDigest.getInstance("MD5");
            do {
                readCount = inputStream.read(buffer);
                if (readCount > 0) {
                    md.update(buffer, 0, readCount);
                }
            } while (readCount != -1);
            byte[] digest = md.digest();
            return String.format("%032x", new BigInteger(1, digest));
        } catch (NoSuchAlgorithmException | IOException e) {
            throw new RuntimeException(e);
        }
    }

    public static String get(String str) {
        return get(str.getBytes(StandardCharsets.UTF_8));
    }

    public static void main(String[] args) {
        System.out.println(get("root:123456"));
    }

}
