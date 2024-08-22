package net.hserver.hplite.domian.bean;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ResDeviceInfo {

    private String deviceId;

    private String desc;

    private boolean online;
    public ResDeviceInfo(String deviceId, String desc, boolean online) {
        this.deviceId = deviceId;
        this.desc = desc;
        this.online = online;
    }
    private MemoryInfo memoryInfo;
}
