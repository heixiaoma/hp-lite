package net.hserver.hplite.service;

import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import net.hserver.hplite.dao.UserCustomDao;
import net.hserver.hplite.domian.entity.UserCustomEntity;

import java.util.Date;

@Bean
public class UserCustomService {

    @Autowired
    private UserCustomDao userCustomDao;


    public Object getAdminPage(Integer page, Integer pageSize) {
        return userCustomDao.selectPage(new Page<>(page, pageSize), new LambdaQueryWrapper<UserCustomEntity>().orderByDesc(UserCustomEntity::getCreateTime));
    }

    public void adminSave(UserCustomEntity userCustomEntity) {
        if (userCustomEntity.getId()==null){
            userCustomEntity.setCreateTime(new Date());
        }
        userCustomDao.insertOrUpdate(userCustomEntity);
    }

    public void adminRemove(Integer id) {
        userCustomDao.deleteById(id);
    }

}
