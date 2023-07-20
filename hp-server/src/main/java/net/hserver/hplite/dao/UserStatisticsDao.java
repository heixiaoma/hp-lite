package net.hserver.hplite.dao;

import cn.hserver.plugin.mybatis.annotation.Mybatis;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import net.hserver.hplite.domian.entity.UserStatisticsEntity;
import org.apache.ibatis.annotations.Update;

@Mybatis
public interface UserStatisticsDao extends BaseMapper<UserStatisticsEntity> {

    @Update("DELETE FROM user_statistics WHERE create_time < DATE('now', '-1 day');")
    int deleteOldData();

}
