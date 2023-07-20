package net.hserver.hplite.queue;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.queue.QueueHandler;
import cn.hserver.core.ioc.annotation.queue.QueueListener;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.handler.CmdServerHandler;
import net.hserver.hplite.message.LocalInnerWear;
import net.hserver.hplite.service.DeviceService;
import net.hserver.hplite.service.UserConfigService;

import java.util.List;
import java.util.stream.Collectors;

@QueueListener(queueName = "CONNECT_EVENT")
@Slf4j
public class ConnectEvent {

    @Autowired
    private UserConfigService userConfigService;

    @Autowired
    private DeviceService deviceService;

    @QueueHandler
    public void connect(String key) {
        if (deviceService.hasKey(key)) {
            List<UserConfigEntity> userConfig = userConfigService.getDeviceConfig(key);
            if (userConfig != null) {
                List<LocalInnerWear> collect = userConfig.stream().map(k -> {
                    LocalInnerWear localInnerWear = new LocalInnerWear();
                    localInnerWear.setServerIp(k.getServerIp());
                    localInnerWear.setServerPort(k.getServerPort());
                    localInnerWear.setLocalIp(k.getLocalIp());
                    localInnerWear.setLocalPort(k.getLocalPort());
                    localInnerWear.setConfigKey(k.getConfigKey());
                    localInnerWear.setConnectType(k.getConnectType());
                    return localInnerWear;
                }).collect(Collectors.toList());
                //给客服端下发穿透数据
                boolean send = CmdServerHandler.send(key, collect);
                log.info("{}-穿透数据下发结果:{}", key, send);
            }
        } else {
            boolean b = CmdServerHandler.sendCloseMsg(key, "设备ID不存在，请检查你的设备ID配置是否正确");
            log.info("未知设备连接，关闭设备:{},结果：{}", key, b);
        }
    }

}
