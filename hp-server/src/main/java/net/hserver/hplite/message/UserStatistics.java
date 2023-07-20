package net.hserver.hplite.message;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.ToString;

import java.util.concurrent.atomic.LongAdder;

@Data
@AllArgsConstructor
@NoArgsConstructor
@ToString
public class UserStatistics {
    private Integer userId;
    private Integer configId;
    private LongAdder download = new LongAdder();
    private LongAdder upload = new LongAdder();
    private int uv;
    private int pv;
    private boolean hasPackageNoExp;
    private boolean hasUpdateFlow;

    public void rest() {
        download.reset();
        upload.reset();
    }
}
