package com.rainlf.mgttbe.infra.util;

import com.rainlf.mgttbe.infra.db.dataobj.MgttConfig;
import com.rainlf.mgttbe.infra.db.repository.MgttConfigRepository;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;

/**
 * 初始化微信配置数据的工具类
 */
@Component
@ConditionalOnProperty(name = "mgtt.data.initialize", havingValue = "true")
public class ConfigInitializer {
    @Autowired
    private MgttConfigRepository mgttConfigRepository;

    @PostConstruct
    public void init() {
        // 定义需要插入的微信配置数据
        Map<String, String> configMap = new HashMap<>();
        configMap.put("wx.app.id", "wx91619e0ec88f17ea");
        configMap.put("wx.app.secret", "9b1dc445c1cafd228bea79e6c6540b86");
        configMap.put("wx.login.url", "https://api.weixin.qq.com/sns/jscode2session");

        // 检查并插入配置数据
        for (Map.Entry<String, String> entry : configMap.entrySet()) {
            String key = entry.getKey();
            String value = entry.getValue();

            // 检查配置是否已存在
            MgttConfig existingConfig = mgttConfigRepository.findByKey(key);
            if (existingConfig == null) {
                // 创建新的配置记录
                MgttConfig newConfig = new MgttConfig();
                newConfig.setKey(key);
                newConfig.setValue(value);
                
                // 保存到数据库
                mgttConfigRepository.save(newConfig);
                System.out.println("已插入配置: " + key + " = " + value);
            } else {
                System.out.println("配置已存在，跳过: " + key);
            }
        }
    }
}