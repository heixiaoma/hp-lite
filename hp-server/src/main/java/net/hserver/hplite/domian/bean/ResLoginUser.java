package net.hserver.hplite.domian.bean;

import lombok.Data;
import net.hserver.hplite.config.AdminInfo;
import net.hserver.hplite.domian.bean.Token;
import net.hserver.hplite.utils.TokenUtil;

@Data
public class ResLoginUser {

    private String token;

    private Long expTime;

    private String email;


    public ResLoginUser(AdminInfo adminInfo) {
        this.token = TokenUtil.genToken(adminInfo.getUsername().hashCode(), Token.Role.ADMIN);
        this.setEmail(adminInfo.getUsername());
        //三天的到期时间
        this.expTime = System.currentTimeMillis() + 86400000 * 3;
    }
}
