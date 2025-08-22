package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import net.hserver.hplite.dao.UserConfigDao;
import net.hserver.hplite.dao.UserStatisticsDao;
import net.hserver.hplite.domian.bean.Token;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.domian.entity.UserStatisticsEntity;
import net.hserver.hplite.utils.TokenUtil;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Bean
public class MonitorService {
    @Autowired
    private UserStatisticsDao userStatisticsDao;
    @Autowired
    private UserConfigDao userConfigDao;
    public Map<Integer, List<UserStatisticsEntity>> getList() {
        Token token = TokenUtil.getToken();
        List<Integer> collect=null;
        if (token.getRole() == Token.Role.CLIENT) {
            List<UserConfigEntity> userConfigEntities = userConfigDao.selectList(
                    new LambdaQueryWrapper<UserConfigEntity>()
                            .eq(token.getRole() == Token.Role.CLIENT, UserConfigEntity::getUserId, token.getUserId())
            );
           collect = userConfigEntities.stream().map(UserConfigEntity::getId).collect(Collectors.toList());
        }
        List<UserStatisticsEntity> userStatisticsEntities = userStatisticsDao.selectList(
                new LambdaQueryWrapper<UserStatisticsEntity>()
                        .in(collect!=null&&token.getRole() == Token.Role.CLIENT, UserStatisticsEntity::getConfigId, collect)
        );
        return userStatisticsEntities.stream().collect(Collectors.groupingBy(UserStatisticsEntity::getConfigId));
    }


}
