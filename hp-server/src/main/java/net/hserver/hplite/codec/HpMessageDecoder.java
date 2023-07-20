package net.hserver.hplite.codec;

import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.ByteToMessageDecoder;
import net.hserver.hplite.message.HpMessageData;

import java.util.List;

/**
 * @author hxm
 */
public class HpMessageDecoder extends ByteToMessageDecoder {
    /**
     * <pre>
     * 协议开始的标准head_data，int类型，占据4个字节.
     * 表示是否解压isUncompress，byte类型，占据1个字节. 1标识要解压 否则不解压
     * 表示数据的长度contentLength，int类型，占据4个字节.
     * </pre>
     */
    public final int BASE_LENGTH = 4 + 4;

    @Override
    protected void decode(ChannelHandlerContext ctx, ByteBuf in, List<Object> out) throws Exception {
        // 可读长度必须大于基本长度
        if (in.readableBytes() >= BASE_LENGTH) {
            in.markReaderIndex();
            int header = in.readInt();
            if (header == 9999) {
                int dataLength = in.readInt();
                if (in.readableBytes() < dataLength) {
                    in.resetReaderIndex();
                } else {
                    byte[] data = new byte[dataLength];
                    in.readBytes(data);
                    out.add(HpMessageData.HpMessage.parseFrom(data));
                }
            }
        }
    }

}
