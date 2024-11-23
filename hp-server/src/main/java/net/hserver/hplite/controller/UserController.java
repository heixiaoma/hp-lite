package net.hserver.hplite.controller;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.annotation.Controller;
import cn.hserver.plugin.web.annotation.POST;
import cn.hserver.plugin.web.interfaces.HttpRequest;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.ReqLoginUser;
import net.hserver.hplite.domian.bean.ResLoginUser;
import net.hserver.hplite.service.UserService;

@Controller(value = "/user/")
@Slf4j
public class UserController {
    @Autowired
    private UserService userService;

    @POST("login")
    public JsonResult login(ReqLoginUser reqLoginUser) {
        try {
            ResLoginUser resLoginUser = userService.loginUser(reqLoginUser);
            if (resLoginUser != null) {
                return JsonResult.ok("登录成功").put("data", resLoginUser);
            } else {
                return JsonResult.error("登录失败");
            }
        } catch (Exception e) {
            log.error(e.getMessage(),e);
            return JsonResult.error(e.getMessage());
        }
    }


}
