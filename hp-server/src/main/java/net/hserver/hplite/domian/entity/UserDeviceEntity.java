package net.hserver.hplite.domian.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("user_device")
public class UserDeviceEntity {

    @TableId(type = IdType.AUTO)
    private Integer id;

    /**
     * 设备key
     */
    private String deviceKey;

    /**
     * 描述
     */
    private String remarks;

}
