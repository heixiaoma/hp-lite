package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import net.hserver.hplite.dao.UserStatisticsDao;
import net.hserver.hplite.domian.entity.UserStatisticsEntity;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Bean
public class MonitorService {
    @Autowired
    private UserStatisticsDao userStatisticsDao;

    public Map<Integer, List<UserStatisticsEntity>> getList() {
        List<UserStatisticsEntity> userStatisticsEntities = userStatisticsDao.selectList(new LambdaQueryWrapper<>());
        return userStatisticsEntities.stream().collect(Collectors.groupingBy(UserStatisticsEntity::getConfigId));
    }


}
