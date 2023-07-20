package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import net.hserver.hplite.config.AdminInfo;
import net.hserver.hplite.domian.bean.ResLoginUser;
import net.hserver.hplite.domian.bean.ReqLoginUser;

@Bean
public class UserService {

    @Autowired
    private AdminInfo adminInfo;


    public boolean notNull(Object obj) {
        if (obj != null && obj.toString().trim().length() > 0) {
            return true;
        }
        return false;
    }


    public ResLoginUser loginUser(ReqLoginUser reqLoginUser, String ipAddress) {
        //校验是否是后台用户
        if (notNull(reqLoginUser.getEmail())
                && notNull(reqLoginUser.getPassword())
                && adminInfo.getUsername().equals(reqLoginUser.getEmail())
                && adminInfo.getPassword().equals(reqLoginUser.getPassword())) {
            return new ResLoginUser(adminInfo);
        }

        return null;
    }


}
