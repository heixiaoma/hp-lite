package net.hserver.hplite.domian.bean;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ReqDeviceInfo {

    private String deviceId;

    private String desc;
}
