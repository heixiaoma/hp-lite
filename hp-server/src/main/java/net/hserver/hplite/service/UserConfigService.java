package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.queue.HServerQueue;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.LambdaUpdateWrapper;
import net.hserver.hplite.config.WebConfig;
import net.hserver.hplite.dao.UserConfigDao;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.utils.CheckUtil;

import java.util.*;

@Bean
public class UserConfigService {

    @Autowired
    private WebConfig webConfig;

    @Autowired
    private UserConfigDao userConfigDao;

    public List<UserConfigEntity> getDeviceConfig(String deviceKey) {
        return userConfigDao.selectList(
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getDeviceKey, deviceKey)
        );
    }


    public List<UserConfigEntity> getConfigList() {
        return userConfigDao.selectList(
                new LambdaQueryWrapper<>()
        );
    }

    public UserConfigEntity getConfig(String configKey) {
        return userConfigDao.selectOne(
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getConfigKey, configKey)
                        .last("limit 1")
        );
    }


    public boolean removeConfig(Integer configId) {
        boolean res = false;
        UserConfigEntity userConfigEntity = userConfigDao.selectById(configId);
        if (userConfigEntity != null) {
            res = userConfigDao.deleteById(configId) > 0;
        }
        if (res) {
            HServerQueue.sendQueue("CONNECT_EVENT", userConfigEntity.getConfigKey());
        }
        return res;
    }

    public void addConfig(UserConfigEntity userConfigEntity) throws RuntimeException {

        if (userConfigEntity.getDeviceKey() == null || userConfigEntity.getDeviceKey().trim().length() == 0) {
            throw new RuntimeException("设备ID未选择");
        }

        if (userConfigEntity.getRemarks() == null || userConfigEntity.getRemarks().trim().length() > 200) {
            throw new RuntimeException("备注不能为空，同时不能超过200字");
        }

        if (userConfigEntity.getConnectType() == null) {
            throw new RuntimeException("穿透协议未选择");
        }

        if (userConfigEntity.getPort() == null) {
            throw new RuntimeException("外网端口未填写");
        }
        if (userConfigEntity.getDomain() == null) {
            throw new RuntimeException("穿透域名未填写");
        }

        if (!CheckUtil.isValidIPAddress(userConfigEntity.getLocalIp())) {
            throw new RuntimeException("内网IP填写不正确");
        }

        if (!CheckUtil.isValidPort(userConfigEntity.getLocalPort())) {
            throw new RuntimeException("内网端口填写不正确");
        }
        //域名和端口配置已经校验了端口和域名的唯一，这里校验端口和域名配置是否被用过
        Long portCount = userConfigDao.selectCount(
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getPort, userConfigEntity.getPort())
        );
        if (portCount > 0 && userConfigEntity.getPort() > 0) {
            throw new RuntimeException("端口已被其他服务使用，请换一个");
        }
        Long domainCount = userConfigDao.selectCount(
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getDomain, userConfigEntity.getDomain())
        );
        if (domainCount > 0) {
            throw new RuntimeException("域名已被其他服务使用，请换一个");
        }
        userConfigEntity.setConfigKey(UUID.randomUUID().toString());
        userConfigEntity.setServerIp(webConfig.getHost());
        userConfigEntity.setServerPort(9090);
        userConfigDao.insert(userConfigEntity);
        HServerQueue.sendQueue("CONNECT_EVENT", userConfigEntity.getDeviceKey());
    }

    public void refConfig(Integer configId) {
        UserConfigEntity userConfigEntity = userConfigDao.selectById(configId);
        if (userConfigEntity != null) {
            userConfigDao.update(null,
                    new LambdaUpdateWrapper<UserConfigEntity>()
                            .eq(UserConfigEntity::getId, userConfigEntity.getId())
                            .set(UserConfigEntity::getStatusMsg, null)
                            .set(UserConfigEntity::getConfigKey, UUID.randomUUID().toString())
            );
            HServerQueue.sendQueue("CONNECT_EVENT", userConfigEntity.getDeviceKey());
        }
    }

    public void pushStatus(String configKey, String msg) {
        userConfigDao.update(null,
                new LambdaUpdateWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getConfigKey, configKey)
                        .set(UserConfigEntity::getStatusMsg, msg)
        );
    }

}
