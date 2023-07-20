package net.hserver.hplite.utils;

import java.text.SimpleDateFormat;
import java.util.Date;

/**
 * @author hxm
 */
public class DateUtil {
    public static String stampToDate(String s) {
        try {
            String res;
            SimpleDateFormat simpleDateFormat = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
            long lt = new Long(s);
            Date date = new Date(lt);
            res = simpleDateFormat.format(date);
            return res;
        } catch (Exception e) {
            return "无时间";
        }
    }

    public static String dateToStamp(Date date) {
        String res;
        SimpleDateFormat simpleDateFormat = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        res = simpleDateFormat.format(date);
        return res;
    }
}
