package net.hserver.hplite.codec;

import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.MessageToByteEncoder;
import net.hserver.hplite.message.HpMessageData;

/**
 * @author hxm
 */
public class HpMessageEncoder extends MessageToByteEncoder<HpMessageData.HpMessage> {

    @Override
    protected void encode(ChannelHandlerContext ctx, HpMessageData.HpMessage msg, ByteBuf out) throws Exception {
        // 1.写入消息的开头的信息标志(int类型)
        out.writeInt(9999);
        byte[] bytes = msg.toByteArray();
        // 2.写入消息的长度(int 类型)
        out.writeInt(bytes.length);
        // 3.写入消息的内容(byte[]类型)
        out.writeBytes(bytes);
    }
}
