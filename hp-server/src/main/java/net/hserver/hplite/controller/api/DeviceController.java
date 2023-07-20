package net.hserver.hplite.controller.api;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.annotation.Controller;
import cn.hserver.plugin.web.annotation.GET;
import cn.hserver.plugin.web.annotation.POST;
import cn.hserver.plugin.web.interfaces.HttpRequest;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.ReqDeviceInfo;
import net.hserver.hplite.service.DeviceService;

@Slf4j
@Controller("/client/device/")
public class DeviceController  {

    @Autowired
    private DeviceService deviceService;

    /**
     * 获取设备列表在线
     *
     * @return
     */
    @GET("list")
    public JsonResult getDeviceList() {
        try {
            return JsonResult.ok().put("data", deviceService.getDeviceList());
        } catch (Exception e) {
            e.printStackTrace();
            return JsonResult.error(e.getMessage());
        }
    }

    /**
     * 添加设备
     *
     * @return
     */
    @POST("add")
    public JsonResult addDevice(HttpRequest request, ReqDeviceInfo reqDeviceInfo) {
        try {
            return JsonResult.ok().put("data", deviceService.addDevice(reqDeviceInfo));
        } catch (Exception e) {
            return JsonResult.error(e.getMessage());
        }
    }

    @GET("remove")
    public JsonResult remove(HttpRequest request,String deviceId) {
        try {
            return JsonResult.ok().put("data", deviceService.remove(deviceId));
        } catch (Exception e) {
            return JsonResult.error(e.getMessage());
        }
    }
}
