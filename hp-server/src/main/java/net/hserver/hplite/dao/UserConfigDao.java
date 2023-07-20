package net.hserver.hplite.dao;

import cn.hserver.plugin.mybatis.annotation.Mybatis;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import net.hserver.hplite.domian.entity.UserConfigEntity;


@Mybatis
public interface UserConfigDao extends BaseMapper<UserConfigEntity> {

}
