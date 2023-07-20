package net.hserver.hplite.config;

import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Configuration;
import cn.hserver.plugin.mybatis.bean.MybatisConfig;
import com.baomidou.mybatisplus.annotation.DbType;
import com.baomidou.mybatisplus.extension.plugins.MybatisPlusInterceptor;
import com.baomidou.mybatisplus.extension.plugins.inner.PaginationInnerInterceptor;
import com.zaxxer.hikari.HikariDataSource;
import net.hserver.hplite.utils.FileUtil;
import org.apache.ibatis.plugin.Interceptor;

import java.io.File;
import java.io.InputStream;

/**
 * @author hxm
 */
@Configuration
public class DbConfig {

    @Bean
    public MybatisConfig sql() {
        String path = System.getProperty("user.dir") + File.separator + "db.db";
        File file = new File(path);
        if (!file.exists()) {
            //进行导出初始化一个sqlite
            InputStream is = DbConfig.class.getResourceAsStream("/db/db.db");
            FileUtil.copyFile(is, path);
        }
        HikariDataSource ds = new HikariDataSource();
        ds.setJdbcUrl( "jdbc:sqlite:" + path);
        ds.setDriverClassName("org.sqlite.JDBC");
        MybatisConfig mybatisConfig = new MybatisConfig();
        mybatisConfig.setMapUnderscoreToCamelCase(true);
        //默认数据源
        mybatisConfig.addDataSource(ds);
        //resource/mapper 全部.xml扫描
        mybatisConfig.setMapperLocations("mapper");
        //分页插件
        mybatisConfig.setPlugins(new Interceptor[]{initInterceptor()});
        return mybatisConfig;

    }
    private Interceptor initInterceptor() {
        //创建mybatis-plus插件对象
        MybatisPlusInterceptor interceptor = new MybatisPlusInterceptor();
        //构建分页插件
        PaginationInnerInterceptor paginationInnerInterceptor = new PaginationInnerInterceptor();
        paginationInnerInterceptor.setDbType(DbType.MYSQL);
        paginationInnerInterceptor.setOverflow(true);
        paginationInnerInterceptor.setMaxLimit(500L);
        interceptor.addInnerInterceptor(paginationInnerInterceptor);
        return interceptor;
    }
}
