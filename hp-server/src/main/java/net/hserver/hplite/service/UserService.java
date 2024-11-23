package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import net.hserver.hplite.config.AdminInfo;
import net.hserver.hplite.dao.UserCustomDao;
import net.hserver.hplite.domian.bean.ResLoginUser;
import net.hserver.hplite.domian.bean.ReqLoginUser;
import net.hserver.hplite.domian.entity.UserCustomEntity;

@Bean
public class UserService {

    @Autowired
    private AdminInfo adminInfo;

    @Autowired
    private UserCustomDao userCustomDao;


    public boolean notNull(Object obj) {
        if (obj != null && obj.toString().trim().length() > 0) {
            return true;
        }
        return false;
    }


    public ResLoginUser loginUser(ReqLoginUser reqLoginUser) {
        //校验是否是后台用户
        if (notNull(reqLoginUser.getEmail())
                && notNull(reqLoginUser.getPassword())
                && adminInfo.getUsername().equals(reqLoginUser.getEmail())
                && adminInfo.getPassword().equals(reqLoginUser.getPassword())) {
            return new ResLoginUser(adminInfo);
        }
        UserCustomEntity userCustomEntity = userCustomDao.selectOne(
                new LambdaQueryWrapper<UserCustomEntity>()
                        .eq(UserCustomEntity::getUsername, reqLoginUser.getEmail())
                        .eq(UserCustomEntity::getPassword, reqLoginUser.getPassword())
                        .last("limit 1")
        );
        if (userCustomEntity!=null){
            return new ResLoginUser(userCustomEntity);
        }
        return null;
    }


}
