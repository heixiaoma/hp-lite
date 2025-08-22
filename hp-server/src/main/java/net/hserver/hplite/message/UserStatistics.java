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
    private Integer configId;
    private LongAdder download = new LongAdder();
    private LongAdder upload = new LongAdder();
    private LongAdder uv=new LongAdder();
    private LongAdder pv=new LongAdder();

    public void rest() {
        uv.reset();
        pv.reset();
        download.reset();
        upload.reset();
    }
}
