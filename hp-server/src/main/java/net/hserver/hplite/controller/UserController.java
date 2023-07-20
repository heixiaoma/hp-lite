package net.hserver.hplite.controller;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.annotation.Controller;
import cn.hserver.plugin.web.annotation.POST;
import cn.hserver.plugin.web.interfaces.HttpRequest;
import net.hserver.hplite.domian.bean.ReqLoginUser;
import net.hserver.hplite.domian.bean.ResLoginUser;
import net.hserver.hplite.service.UserService;

@Controller(value = "/user/")
public class UserController {
    @Autowired
    private UserService userService;

    @POST("login")
    public JsonResult login(HttpRequest request, ReqLoginUser reqLoginUser) {
        try {
            ResLoginUser resLoginUser = userService.loginUser(reqLoginUser, request.getIpAddress());
            if (resLoginUser != null) {
                return JsonResult.ok("登录成功").put("data", resLoginUser);
            } else {
                return JsonResult.error("登录失败");
            }
        } catch (Exception e) {
            e.printStackTrace();
            return JsonResult.error(e.getMessage());
        }
    }


}
