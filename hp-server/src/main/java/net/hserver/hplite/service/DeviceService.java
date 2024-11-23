package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.LambdaUpdateWrapper;
import net.hserver.hplite.dao.UserConfigDao;
import net.hserver.hplite.dao.UserDeviceDao;
import net.hserver.hplite.domian.bean.OnlineInfo;
import net.hserver.hplite.domian.bean.ReqDeviceInfo;
import net.hserver.hplite.domian.bean.ResDeviceInfo;
import net.hserver.hplite.domian.bean.ResUserKey;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.domian.entity.UserDeviceEntity;
import net.hserver.hplite.handler.cmd.CmdServerHandler;
import net.hserver.hplite.utils.CheckUtil;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Bean
public class DeviceService {


    @Autowired
    private UserConfigDao userConfigDao;

    @Autowired
    private UserDeviceDao deviceDao;
    public List<UserDeviceEntity> getUserDeviceList() {
        return deviceDao.selectList(new LambdaQueryWrapper<>());
    }
    /**
     * 查用户的key和是否在线
     *
     * @return
     */
    public List<ResDeviceInfo> getDeviceList() {
        List<UserDeviceEntity> userDeviceList = getUserDeviceList();
        List<ResDeviceInfo> deviceInfos = new ArrayList<>();
        if (userDeviceList.isEmpty()) {
            return deviceInfos;
        }
        for (UserDeviceEntity deviceEntity : userDeviceList) {
            OnlineInfo onlineKey = CmdServerHandler.getOnlineKey(deviceEntity.getDeviceKey());
            ResDeviceInfo resDeviceInfo = new ResDeviceInfo(deviceEntity.getDeviceKey(), deviceEntity.getRemarks(), onlineKey != null);
            if (onlineKey != null && onlineKey.getMemoryInfo() != null) {
                resDeviceInfo.setMemoryInfo(onlineKey.getMemoryInfo());
            }
            deviceInfos.add(resDeviceInfo);
        }
        return deviceInfos;
    }

    public List<ResUserKey> getDeviceKey() {
        List<UserDeviceEntity> userDeviceEntities = deviceDao.selectList(new LambdaQueryWrapper<>());
        return userDeviceEntities.stream().map(k -> {
            ResUserKey resUserKey = new ResUserKey();
            resUserKey.setKey(k.getDeviceKey());
            boolean b = CmdServerHandler.getOnlineKey(k.getDeviceKey())!=null;
            resUserKey.setDesc((b ? "在线" : "离线") + "-" + k.getRemarks());
            return resUserKey;
        }).collect(Collectors.toList());
    }


    public boolean updateDevice(ReqDeviceInfo reqDeviceInfo) {
        String desc = reqDeviceInfo.getDesc();
        if (desc == null || desc.trim().length() == 0) {
            throw new RuntimeException("设备备注不能为空");
        }
        String deviceId = reqDeviceInfo.getDeviceId();
        if (deviceId == null || deviceId.trim().length() != 32 || !CheckUtil.checkDomain(deviceId)) {
            throw new RuntimeException("设备编号，不符合规范");
        }
       return deviceDao.update(null,
                new LambdaUpdateWrapper<UserDeviceEntity>()
                        .eq(UserDeviceEntity::getDeviceKey,reqDeviceInfo.getDeviceId())
                        .set(UserDeviceEntity::getRemarks,reqDeviceInfo.getDesc())
        )>0;
    }




    public boolean addDevice(ReqDeviceInfo reqDeviceInfo) {

        String desc = reqDeviceInfo.getDesc();
        if (desc == null || desc.trim().length() == 0) {
            throw new RuntimeException("设备备注不能为空");
        }
        String deviceId = reqDeviceInfo.getDeviceId();
        if (deviceId == null || deviceId.trim().length() != 32 || !CheckUtil.checkDomain(deviceId)) {
            throw new RuntimeException("设备编号，不符合规范");
        }
        Long aLong = deviceDao.selectCount(
                new LambdaQueryWrapper<UserDeviceEntity>()
                        .eq(UserDeviceEntity::getDeviceKey, reqDeviceInfo.getDeviceId())
        );

        if (aLong > 0) {
            throw new RuntimeException("设备编号，已存在");
        }
        UserDeviceEntity userDeviceEntity = new UserDeviceEntity();
        userDeviceEntity.setDeviceKey(reqDeviceInfo.getDeviceId());
        userDeviceEntity.setRemarks(reqDeviceInfo.getDesc());
        deviceDao.insert(userDeviceEntity);
        return true;
    }

    public boolean remove(String deviceId) {
        if (deviceId == null || deviceId.trim().length() != 32 || !CheckUtil.checkDomain(deviceId)) {
            throw new RuntimeException("设备编号，不符合规范");
        }
        if (userConfigDao.selectCount(
                new LambdaQueryWrapper<UserConfigEntity>()
                        .eq(UserConfigEntity::getDeviceKey, deviceId)
        ) > 0) {
            throw new RuntimeException("设备被占用，请停止后在删除");
        }
        return deviceDao.delete(
                new LambdaQueryWrapper<UserDeviceEntity>()
                        .eq(UserDeviceEntity::getDeviceKey, deviceId)) > 0;
    }

    public boolean hasKey(String key) {
        return deviceDao.selectCount(
                new LambdaQueryWrapper<UserDeviceEntity>()
                        .eq(UserDeviceEntity::getDeviceKey, key)
        ) > 0;
    }

    public Object stop(String deviceId) {
        return CmdServerHandler.sendCloseMsg(deviceId, "强制停止");
    }
}
