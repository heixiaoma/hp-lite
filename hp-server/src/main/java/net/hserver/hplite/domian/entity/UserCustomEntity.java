package net.hserver.hplite.domian.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("user_custom")
public class UserCustomEntity{
    @TableId(type = IdType.AUTO)
    private Integer id;

    private String username;

    private String password;

    private String desc;

    private Date createTime;

}
