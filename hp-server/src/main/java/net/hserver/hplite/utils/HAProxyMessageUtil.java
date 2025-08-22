package net.hserver.hplite.utils;

import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.handler.codec.haproxy.*;
import io.netty.util.CharsetUtil;
import io.netty.util.NetUtil;

import java.lang.reflect.Method;
import java.util.List;

public class HAProxyMessageUtil {
    static final byte[] TEXT_PREFIX = new byte[]{80, 82, 79, 88, 89};
    static final byte[] BINARY_PREFIX = new byte[]{13, 10, 13, 10, 0, 13, 10, 81, 85, 73, 84, 10};


    public static byte[] encodeBytes( HAProxyMessage msg) {
        ByteBuf byteBuf = encodeByteBuf(msg);
        byte[] bytes = io.netty.buffer.ByteBufUtil.getBytes(byteBuf);
        byteBuf.release();
        return bytes;
    }

    public static ByteBuf encodeByteBuf( HAProxyMessage msg) {
        ByteBuf buffer = Unpooled.buffer();
        switch (msg.protocolVersion()) {
            case V1:
                encodeV1(msg, buffer);
                break;
            case V2:
                encodeV2(msg, buffer);
                break;
            default:
                throw new HAProxyProtocolException("Unsupported version: " + msg.protocolVersion());
        }
        return buffer;
    }

    private static void encodeV1(HAProxyMessage msg, ByteBuf out) {
        out.writeBytes(TEXT_PREFIX);
        out.writeByte(32);
        out.writeCharSequence(msg.proxiedProtocol().name(), CharsetUtil.US_ASCII);
        out.writeByte(32);
        out.writeCharSequence(msg.sourceAddress(), CharsetUtil.US_ASCII);
        out.writeByte(32);
        out.writeCharSequence(msg.destinationAddress(), CharsetUtil.US_ASCII);
        out.writeByte(32);
        out.writeCharSequence(String.valueOf(msg.sourcePort()), CharsetUtil.US_ASCII);
        out.writeByte(32);
        out.writeCharSequence(String.valueOf(msg.destinationPort()), CharsetUtil.US_ASCII);
        out.writeByte(13);
        out.writeByte(10);
    }

    private static void encodeV2(HAProxyMessage msg, ByteBuf out) {
        out.writeBytes(BINARY_PREFIX);
        out.writeByte(32 | msg.command().byteValue());
        out.writeByte(msg.proxiedProtocol().byteValue());
        int tlvNumBytes = 0;
        try {
            Method tlvNumBytesMethod = msg.getClass().getDeclaredMethod("tlvNumBytes");
            tlvNumBytes = (int) tlvNumBytesMethod.invoke(msg);
        } catch (Exception ignored) {
        }
        switch (msg.proxiedProtocol().addressFamily()) {
            case AF_IPv4:
            case AF_IPv6:
                byte[] srcAddrBytes = NetUtil.createByteArrayFromIpAddressString(msg.sourceAddress());
                byte[] dstAddrBytes = NetUtil.createByteArrayFromIpAddressString(msg.destinationAddress());


                out.writeShort(srcAddrBytes.length + dstAddrBytes.length + 4 + tlvNumBytes);
                out.writeBytes(srcAddrBytes);
                out.writeBytes(dstAddrBytes);
                out.writeShort(msg.sourcePort());
                out.writeShort(msg.destinationPort());
                encodeTlvs(msg.tlvs(), out);
                break;
            case AF_UNIX:
                out.writeShort(216 + tlvNumBytes);
                int srcAddrBytesWritten = out.writeCharSequence(msg.sourceAddress(), CharsetUtil.US_ASCII);
                out.writeZero(108 - srcAddrBytesWritten);
                int dstAddrBytesWritten = out.writeCharSequence(msg.destinationAddress(), CharsetUtil.US_ASCII);
                out.writeZero(108 - dstAddrBytesWritten);
                encodeTlvs(msg.tlvs(), out);
                break;
            case AF_UNSPEC:
                out.writeShort(0);
                break;
            default:
                throw new HAProxyProtocolException("unexpected addrFamily");
        }

    }

    private static void encodeTlv(HAProxyTLV haProxyTLV, ByteBuf out) {
        if (haProxyTLV instanceof HAProxySSLTLV) {
            HAProxySSLTLV ssltlv = (HAProxySSLTLV) haProxyTLV;
            int contentNumBytes = 0;
            try {
                Method contentNumBytesMethod = ssltlv.getClass().getDeclaredMethod("contentNumBytes");
                contentNumBytes = (int) contentNumBytesMethod.invoke(ssltlv);
            } catch (Exception e) {
                e.printStackTrace();
            }
            out.writeByte(haProxyTLV.typeByteValue());
            out.writeShort(contentNumBytes);
            out.writeByte(ssltlv.client());
            out.writeInt(ssltlv.verify());
            encodeTlvs(ssltlv.encapsulatedTLVs(), out);
        } else {
            out.writeByte(haProxyTLV.typeByteValue());
            ByteBuf value = haProxyTLV.content();
            int readableBytes = value.readableBytes();
            out.writeShort(readableBytes);
            out.writeBytes(value.readSlice(readableBytes));
        }

    }

    private static void encodeTlvs(List<HAProxyTLV> haProxyTLVs, ByteBuf out) {
        for (int i = 0; i < haProxyTLVs.size(); ++i) {
            encodeTlv(haProxyTLVs.get(i), out);
        }
    }

    public static void main(String[] args) throws Exception{
        HAProxyMessage message = new HAProxyMessage(
                HAProxyProtocolVersion.valueOf("V1"), HAProxyCommand.PROXY, HAProxyProxiedProtocol.TCP4,
                "127.0.0.1", "127.0.0.1",8080,8080);
        byte[] encode = encodeBytes(message);
        System.out.println(new String(encode));
    }

}
