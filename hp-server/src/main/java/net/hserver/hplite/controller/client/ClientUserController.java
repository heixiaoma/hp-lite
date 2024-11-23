package net.hserver.hplite.controller.client;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.annotation.Controller;
import cn.hserver.plugin.web.annotation.GET;
import cn.hserver.plugin.web.annotation.POST;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.Token;
import net.hserver.hplite.domian.entity.UserCustomEntity;
import net.hserver.hplite.service.UserCustomService;
import net.hserver.hplite.utils.TokenUtil;


@Slf4j
@Controller("/client/user/")
public class ClientUserController {
    @Autowired
    private UserCustomService userCustomService;

    @POST("saveUser")
    public JsonResult add(UserCustomEntity userCustomEntity) {
        if (TokenUtil.getToken().getRole()!= Token.Role.ADMIN) {
           return JsonResult.error(-2, "用户权限校验失败");
        }
        userCustomService.adminSave(userCustomEntity);
        return JsonResult.ok();
    }

    @GET("list")
    public JsonResult list(Integer page, Integer pageSize) {
        if (TokenUtil.getToken().getRole()!= Token.Role.ADMIN) {
            return JsonResult.error(-2, "用户权限校验失败");
        }
        return JsonResult.ok().put("data", userCustomService.getAdminPage(page, pageSize));
    }

    @GET("removeUser")
    public JsonResult remove(Integer id) {
        if (TokenUtil.getToken().getRole()!= Token.Role.ADMIN) {
            return JsonResult.error(-2, "用户权限校验失败");
        }
        userCustomService.adminRemove(id);
        return JsonResult.ok();
    }
}
