package net.hserver.hplite.codec;

import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.MessageToByteEncoder;
import net.hserver.hplite.message.CmdMessageData;

/**
 * @author hxm
 */
public class CmdMessageEncoder extends MessageToByteEncoder<CmdMessageData.CmdMessage> {

    @Override
    protected void encode(ChannelHandlerContext ctx, CmdMessageData.CmdMessage msg, ByteBuf out) throws Exception {
        // 1.写入消息的开头的信息标志(int类型)
        out.writeInt(6666);
        byte[] bytes = msg.toByteArray();
        // 2.写入消息的长度(int 类型)
        out.writeInt(bytes.length);
        // 3.写入消息的内容(byte[]类型)
        out.writeBytes(bytes);
    }
}
