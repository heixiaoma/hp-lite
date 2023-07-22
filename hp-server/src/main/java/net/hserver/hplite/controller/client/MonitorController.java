package net.hserver.hplite.controller.client;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.annotation.Controller;
import cn.hserver.plugin.web.annotation.GET;
import cn.hserver.plugin.web.interfaces.HttpRequest;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.service.MonitorService;

@Slf4j
@Controller("/client/monitor/")
public class MonitorController  {

    @Autowired
    private MonitorService monitorService;

    @GET("list")
    public JsonResult list(HttpRequest request) {
        return JsonResult.ok().put("data", monitorService.getList());
    }

}
