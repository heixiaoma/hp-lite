package net.hserver.hplite.utils;

import cn.hutool.core.codec.Base64;
import cn.hutool.crypto.symmetric.SymmetricAlgorithm;
import cn.hutool.crypto.symmetric.SymmetricCrypto;
import net.hserver.hplite.domian.bean.Token;

import java.nio.charset.StandardCharsets;

/**
 * token工具
 */
public class TokenUtil {

    //随机生成密钥
    private final static byte[] key = "hp_pro_token_key".getBytes(StandardCharsets.UTF_8);

    //构建
    private static final SymmetricCrypto aes = new SymmetricCrypto(SymmetricAlgorithm.AES, key);

    public static String genToken(Integer userId, Token.Role role) {
        //按自己想法给他搞一个token
        String userData = userId +
                ":" +
                role +
                ":" +
                System.currentTimeMillis();
        String encrypt = aes.encryptHex(userData);
        return Base64.encode(encrypt);
    }


    public static Token getToken(String token) {
        try {
            String decode = Base64.decodeStr(token);
            String decrypt = aes.decryptStr(decode);
            String[] split = decrypt.split(":");
            return new Token(Integer.parseInt(split[0]), Token.Role.valueOf(split[1]), Long.parseLong(split[2]));
        } catch (Exception e) {
            return null;
        }
    }

}
