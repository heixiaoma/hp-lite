package net.hserver.hplite.domian.bean;

import lombok.Data;

@Data
public class MemoryInfo {
    private long total;
    private long useMem;
    private double cpuRate;
    private long hpTotalMem;
    private long hpUseMem;
}
