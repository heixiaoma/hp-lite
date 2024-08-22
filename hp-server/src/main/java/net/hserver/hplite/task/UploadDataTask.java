package net.hserver.hplite.task;


import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Task;
import cn.hserver.plugin.web.context.WebConfig;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.config.TunnelConfig;
import net.hserver.hplite.dao.UserStatisticsDao;
import net.hserver.hplite.domian.entity.UserStatisticsEntity;
import net.hserver.hplite.message.DataStatistics;
import net.hserver.hplite.utils.StatisticsUtil;

import java.util.Date;
import java.util.List;

@Bean
@Slf4j
public class UploadDataTask {

    @Autowired
    private UserStatisticsDao userStatisticsDao;

    @Task(name = "removeData", time = "0 */30 * * * ?")
    public void removeData() {
        int i = userStatisticsDao.deleteOldData();
        log.info("删除统计数：{} ", i);
    }

    @Task(name = "每分钟上报数据", time = "0 */1 * * * ?")
    public void serverData() {
        List<DataStatistics> statistics = StatisticsUtil.getStatistics();
        for (DataStatistics dataStatistic : statistics) {
            UserStatisticsEntity userStatisticsEntity = new UserStatisticsEntity();
            userStatisticsEntity.setConfigId(dataStatistic.getConfigId());
            userStatisticsEntity.setDownload(dataStatistic.getDownload());
            userStatisticsEntity.setUpload(dataStatistic.getUpload());
            userStatisticsEntity.setTime(dataStatistic.getTime());
            userStatisticsEntity.setPv(dataStatistic.getPv());
            userStatisticsEntity.setUv(dataStatistic.getUv());
            userStatisticsEntity.setCreateTime(new Date());
            userStatisticsDao.insert(userStatisticsEntity);
        }
    }
}
