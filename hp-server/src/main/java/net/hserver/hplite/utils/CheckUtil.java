package net.hserver.hplite.utils;



import java.util.regex.Pattern;

public class CheckUtil {
    public static boolean checkUsername(String str) {
        return str.endsWith("@qq.com");
    }

    public static boolean isValidEmail(String email) {
        // 定义电子邮件的正则表达式模式
        String emailRegex = "^[a-zA-Z0-9_+&*-]+(?:\\.[a-zA-Z0-9_+&*-]+)*@(?:[a-zA-Z0-9-]+\\.)+[a-zA-Z]{2,7}$";
        Pattern pattern = Pattern.compile(emailRegex); // 编译正则表达式为模式
        return pattern.matcher(email).matches(); // 使用模式匹配邮箱字符串
    }

    public static boolean isValidPassword(String password) {
        return password.trim().length() >= 6;
    }

    public static boolean checkDomain(String str) {
        String regex = "^[a-z0-9]+$";
        return str.matches(regex);
    }


    /**
     * 校验是否是合规的IP
     *
     * @param ipAddress
     * @return
     */
    public static boolean isValidIPAddress(String ipAddress) {
        if (ipAddress == null) {
            return false;
        }
        // 使用正则表达式匹配IP地址的格式
        String pattern = "^((\\d{1,2}|1\\d{2}|2[0-4]\\d|25[0-5])\\.){3}(\\d{1,2}|1\\d{2}|2[0-4]\\d|25[0-5])$";
        return ipAddress.matches(pattern);
    }

    /**
     * 校验端口是否合格
     *
     * @param port
     * @return
     */
    public static boolean isValidPort(Integer port) {
        if (port == null) {
            return false;
        }
        return port >= 0 && port <= 65535;
    }

}
