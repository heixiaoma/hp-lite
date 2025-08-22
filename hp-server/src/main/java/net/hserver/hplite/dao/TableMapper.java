package net.hserver.hplite.dao;

import cn.hserver.plugin.mybatis.annotation.Mybatis;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import org.apache.ibatis.annotations.Update;

import java.util.Date;


@Mybatis
public interface TableMapper  {

    @Update("CREATE TABLE IF NOT EXISTS \"user_statistics\" (\n" +
            "  \"id\" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\n" +
            "  \"config_id\" INTEGER,\n" +
            "  \"download\" TEXT,\n" +
            "  \"upload\" TEXT,\n" +
            "  \"uv\" TEXT,\n" +
            "  \"pv\" TEXT,\n" +
            "  \"time\" integer,\n" +
            "  \"create_time\" DATE\n" +
            ");")
    int createTableUserStatistics();

    @Update("CREATE TABLE IF NOT EXISTS \"user_custom\" (\n" +
            "  \"id\" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\n" +
            "  \"username\" TEXT,\n" +
            "  \"password\" TEXT,\n" +
            "  \"desc\" TEXT,\n" +
            "  \"create_time\" DATE\n" +
            ");")
    int createTableUserCustom();

    @Update("CREATE TABLE IF NOT EXISTS \"user_device\" (\n" +
            "  \"id\" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\n" +
            "  \"user_id\" INTEGER,\n" +
            "  \"device_key\" TEXT,\n" +
            "  \"remarks\" TEXT\n" +
            ");")
    int createTableUserDevice();
    @Update("CREATE TABLE IF NOT EXISTS \"user_config\" (\n" +
            "  \"id\" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,\n" +
            "  \"user_id\" INTEGER,\n" +
            "  \"config_key\" TEXT,\n" +
            "  \"device_key\" TEXT,\n" +
            "  \"server_ip\" TEXT,\n" +
            "  \"server_port\" TEXT,\n" +
            "  \"local_ip\" TEXT,\n" +
            "  \"local_port\" TEXT,\n" +
            "  \"connect_type\" TEXT,\n" +
            "  \"remarks\" TEXT,\n" +
            "  \"port\" TEXT,\n" +
            "  \"domain\" TEXT,\n" +
            "  \"status_msg\" TEXT,\n" +
            "  \"certificate_key\" TEXT,\n" +
            "  \"certificate_content\" TEXT,\n" +
            "  \"proxy_version\" TEXT\n" +
            ");")
    int createTableUserConfig();

}
