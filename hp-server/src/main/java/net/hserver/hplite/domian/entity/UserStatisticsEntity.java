package net.hserver.hplite.domian.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Date;

/**
 * 用户的套餐
 */
@Data
@TableName("user_statistics")
@NoArgsConstructor
public class UserStatisticsEntity {

    @TableId(type = IdType.AUTO)
    private Integer id;

    /**
     * 套餐ID
     */
    private Integer configId;

    /**
     * 下载量
     */
    private Long download;

    /**
     * 上传量
     */
    private Long upload;

    /**
     * uv
     */
    private Integer uv;

    /**
     * pv
     */
    private Integer pv;

    /**
     * 时间
     */
    private Long time;

    /**
     * 创建时间
     */
    private Date createTime;

}
