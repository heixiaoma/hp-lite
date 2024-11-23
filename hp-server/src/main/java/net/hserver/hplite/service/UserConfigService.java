package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.queue.HServerQueue;
import cn.hutool.core.util.StrUtil;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.LambdaUpdateWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import net.hserver.hplite.config.TunnelConfig;
import net.hserver.hplite.dao.UserConfigDao;
import net.hserver.hplite.domian.bean.Token;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.utils.CheckUtil;
import net.hserver.hplite.utils.SSLUtil;
import net.hserver.hplite.utils.TokenUtil;

import java.net.InetSocketAddress;
import java.net.SocketAddress;
import java.util.*;

@Bean
public class UserConfigService {

    @Autowired
    private TunnelConfig tunnelConfig;

    @Autowired
    private UserConfigDao userConfigDao;

    public List<UserConfigEntity> getDeviceConfig(String deviceKey) {
        return userConfigDao.selectList(
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getDeviceKey, deviceKey)
        );
    }


    public Page<UserConfigEntity> getConfigList(Integer page, Integer pageSize) {
        Token token = TokenUtil.getToken();
        return userConfigDao.selectPage(
                new Page<>(page,pageSize),
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(token.getRole()== Token.Role.CLIENT,UserConfigEntity::getUserId,token.getUserId())
                        .orderByDesc(UserConfigEntity::getId)
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

        if (!CheckUtil.isValidIPAddress(userConfigEntity.getLocalIp())) {
            throw new RuntimeException("内网IP填写不正确");
        }

        if (!CheckUtil.isValidPort(userConfigEntity.getLocalPort())) {
            throw new RuntimeException("内网端口填写不正确");
        }
        if (StrUtil.isNotEmpty(userConfigEntity.getCertificateKey()) || StrUtil.isNotEmpty(userConfigEntity.getCertificateContent())) {
            if (SSLUtil.buildSSLContext(userConfigEntity.getCertificateKey(), userConfigEntity.getCertificateContent(), null) == null) {
                throw new RuntimeException("SSL证书无效");
            }
        }
        //域名和端口配置已经校验了端口和域名的唯一，这里校验端口和域名配置是否被用过

        if (userConfigEntity.getId()==null) {
            Long portCount = userConfigDao.selectCount(
                    new LambdaQueryWrapper<UserConfigEntity>()
                            .eq(UserConfigEntity::getPort, userConfigEntity.getPort())
            );
            if (portCount > 0 && userConfigEntity.getPort() > 0) {
                throw new RuntimeException("外网端口已被其他服务使用，请换一个");
            }
            Long domainCount = userConfigDao.selectCount(
                    new LambdaQueryWrapper<UserConfigEntity>()
                            .isNotNull(UserConfigEntity::getDomain)
                            .eq(UserConfigEntity::getDomain, userConfigEntity.getDomain())
            );
            if (domainCount > 0) {
                throw new RuntimeException("外网域名已被其他服务使用，请换一个");
            }
        }
        userConfigEntity.setConfigKey(UUID.randomUUID().toString());
        userConfigEntity.setUserId(TokenUtil.getToken().getUserId());
        userConfigEntity.setServerIp(tunnelConfig.getIp());
        userConfigEntity.setServerPort(tunnelConfig.getPort());
        if (userConfigEntity.getId()!=null){
            userConfigDao.updateById(userConfigEntity);
        }else {
            userConfigDao.insert(userConfigEntity);
        }
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
