package net.hserver.hplite.utils;

import io.netty.handler.ssl.OpenSsl;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.handler.ssl.SslProvider;
import lombok.extern.slf4j.Slf4j;

import java.io.ByteArrayInputStream;

@Slf4j
public class SSLUtil {

    private static SslProvider defaultSslProvider() {
        return OpenSsl.isAvailable() ? SslProvider.OPENSSL : SslProvider.JDK;
    }

    public static SslContext buildSSLContext(String keyContent, String certContent, String pwd) {
        try (
                ByteArrayInputStream pk8Input = new ByteArrayInputStream(keyContent.getBytes());
                ByteArrayInputStream certInputStream = new ByteArrayInputStream(certContent.getBytes());
        ) {
            SslContextBuilder sslContext = SslContextBuilder.forServer(certInputStream, pk8Input, pwd).sslProvider(defaultSslProvider());
            return sslContext.build();
        } catch (Exception e) {
            log.error(e.getMessage(),e);
            return null;
        }
    }

}
