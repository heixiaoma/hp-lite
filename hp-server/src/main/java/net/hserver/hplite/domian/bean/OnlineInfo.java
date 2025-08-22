package net.hserver.hplite.domian.bean;

import io.netty.channel.ChannelHandlerContext;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class OnlineInfo {
    ChannelHandlerContext ctx;
    MemoryInfo memoryInfo;

}
