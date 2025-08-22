package net.hserver.hplite.controller.client;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.annotation.Controller;
import cn.hserver.plugin.web.annotation.GET;
import cn.hserver.plugin.web.annotation.POST;
import cn.hserver.plugin.web.interfaces.HttpRequest;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.service.DeviceService;
import net.hserver.hplite.service.UserConfigService;

@Slf4j
@Controller("/client/config/")
public class ConfigController {
    @Autowired
    private UserConfigService userConfigService;

    @Autowired
    private DeviceService deviceService;

    @GET("getDeviceKey")
    public JsonResult getUserKey(HttpRequest request) {
        return JsonResult.ok().put("data", deviceService.getDeviceKey());
    }


    @GET("getConfigList")
    public JsonResult getConfigList(Integer current,Integer pageSize) {
        if (current==null){
            current=1;
        }
        if (pageSize==null||pageSize>100){
            pageSize=10;
        }
        return JsonResult.ok().put("data", userConfigService.getConfigList(current,pageSize));
    }


    /**
     * 删除配置
     *
     * @param configId
     * @return
     */
    @GET("removeConfig")
    public JsonResult removeConfig(Integer configId) {
        try {
            return JsonResult.ok().put("data", userConfigService.removeConfig(configId));
        } catch (Exception e) {
            return JsonResult.error(e.getMessage());
        }
    }

    @GET("refConfig")
    public JsonResult refConfig(HttpRequest request, Integer configId) {
        userConfigService.refConfig( configId);
        return JsonResult.ok();
    }

    @POST("addConfig")
    public JsonResult addConfig( UserConfigEntity userConfigEntity) {
        try {
            userConfigService.addConfig( userConfigEntity);
            return JsonResult.ok();
        } catch (Exception e) {
            return JsonResult.error(e.getMessage());
        }
    }

}
