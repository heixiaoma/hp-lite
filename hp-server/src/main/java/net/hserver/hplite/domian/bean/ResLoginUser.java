package net.hserver.hplite.domian.bean;

import lombok.Data;
import net.hserver.hplite.config.AdminInfo;
import net.hserver.hplite.domian.bean.Token;
import net.hserver.hplite.domian.entity.UserCustomEntity;
import net.hserver.hplite.utils.TokenUtil;

@Data
public class ResLoginUser {

    private String token;

    private Long expTime;

    private String email;
    private Token.Role role;


    public ResLoginUser(AdminInfo adminInfo) {
        this.token = TokenUtil.genToken(-1, Token.Role.ADMIN);
        this.setEmail(adminInfo.getUsername());
        //三天的到期时间
        this.expTime = System.currentTimeMillis() + 86400000 * 3;
        this.role= Token.Role.ADMIN;
    }

    public ResLoginUser(UserCustomEntity userCustomEntity) {
        this.token = TokenUtil.genToken(userCustomEntity.getId(), Token.Role.CLIENT);
        this.setEmail(userCustomEntity.getUsername());
        //三天的到期时间
        this.expTime = System.currentTimeMillis() + 86400000 * 3;
        this.role= Token.Role.CLIENT;
    }


}
