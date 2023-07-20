package net.hserver.hplite.message;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.ToString;


@Data
@NoArgsConstructor
@ToString
public class DataStatistics {
    private Integer userId;
    private Integer configId;
    private long download;
    private long upload;
    private int uv;
    private int pv;
    private boolean hasPackageNoExp;
    private boolean hasUpdateFlow;
    private long time;

    public DataStatistics(UserStatistics userStatistics) {
        this.setUserId(userStatistics.getUserId());
        this.setConfigId(userStatistics.getConfigId());
        this.setDownload(userStatistics.getDownload().longValue());
        this.setUpload(userStatistics.getUpload().longValue());
        this.setUv(userStatistics.getUv());
        this.setPv(userStatistics.getPv());
        this.setHasPackageNoExp(userStatistics.isHasPackageNoExp());
        this.setHasUpdateFlow(userStatistics.isHasUpdateFlow());
        this.setTime(System.currentTimeMillis());
    }
}
