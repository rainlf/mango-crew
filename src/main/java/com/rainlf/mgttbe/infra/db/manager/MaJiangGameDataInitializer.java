package com.rainlf.mgttbe.infra.db.manager;

import com.rainlf.mgttbe.infra.db.dataobj.MaJiangGameDO;
import com.rainlf.mgttbe.infra.db.dataobj.MaJiangGameItemDO;
import com.rainlf.mgttbe.infra.db.repository.MaJiangGameItemRepository;
import com.rainlf.mgttbe.infra.db.repository.MaJiangGameRepository;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Random;

@Component
@ConditionalOnProperty(name = "mgtt.data.initialize", havingValue = "true")
public class MaJiangGameDataInitializer implements CommandLineRunner {

    private final MaJiangGameRepository maJiangGameRepository;
    private final MaJiangGameItemRepository maJiangGameItemRepository;
    private final Random random = new Random();

    public MaJiangGameDataInitializer(MaJiangGameRepository maJiangGameRepository, MaJiangGameItemRepository maJiangGameItemRepository) {
        this.maJiangGameRepository = maJiangGameRepository;
        this.maJiangGameItemRepository = maJiangGameItemRepository;
    }

    @Override
    @Transactional
    public void run(String... args) throws Exception {
        // 检查是否已有数据
        if (maJiangGameRepository.count() == 0) {
            System.out.println("开始初始化测试数据...");
            
            // 创建测试用户ID (简化处理，假设用户ID为1,2,3,4)
            List<Integer> userIds = Arrays.asList(1, 2, 3, 4);
            
            // 创建20条游戏记录
            List<MaJiangGameDO> games = new ArrayList<>();
            for (int i = 0; i < 20; i++) {
                MaJiangGameDO game = createGame(userIds);
                games.add(game);
            }
            
            // 保存游戏记录
            List<MaJiangGameDO> savedGames = maJiangGameRepository.saveAll(games);
            
            // 为每个游戏创建游戏项目记录
            List<MaJiangGameItemDO> gameItems = new ArrayList<>();
            for (MaJiangGameDO game : savedGames) {
                gameItems.addAll(createGameItems(game));
            }
            
            // 保存游戏项目记录
            maJiangGameItemRepository.saveAll(gameItems);
            
            System.out.println("测试数据初始化完成：创建了" + savedGames.size() + "条游戏记录和" + gameItems.size() + "条游戏项目记录");
        }
    }

    private MaJiangGameDO createGame(List<Integer> userIds) {
        MaJiangGameDO game = new MaJiangGameDO();
        
        // 随机分配4个不同的玩家
        List<Integer> shuffledUsers = new ArrayList<>(userIds);
        java.util.Collections.shuffle(shuffledUsers);
        game.setPlayer1(shuffledUsers.get(0));
        game.setPlayer2(shuffledUsers.get(1));
        game.setPlayer3(shuffledUsers.get(2));
        game.setPlayer4(shuffledUsers.get(3));
        
        // 随机游戏类型
        game.setType(random.nextInt(3) + 1);
        
        // 移除不存在的score属性设置
        
        // 随机创建时间（最近30天内）
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime createdTime = now.minus(random.nextInt(30), ChronoUnit.DAYS);
        game.setCreatedTime(createdTime);
        game.setUpdatedTime(createdTime);
        
        // 未删除
        game.setIsDeleted(0);
        
        return game;
    }

    private List<MaJiangGameItemDO> createGameItems(MaJiangGameDO game) {
        List<MaJiangGameItemDO> items = new ArrayList<>();
        
        // 为每个玩家创建一条游戏项目记录
        // 临时修改：使用随机分数代替不存在的score属性
        items.add(createGameItem(game.getId(), game.getPlayer1(), random.nextInt(1000)));
        items.add(createGameItem(game.getId(), game.getPlayer2(), random.nextInt(1000)));
        items.add(createGameItem(game.getId(), game.getPlayer3(), random.nextInt(1000)));
        items.add(createGameItem(game.getId(), game.getPlayer4(), random.nextInt(1000)));
        
        return items;
    }

    private MaJiangGameItemDO createGameItem(Integer gameId, Integer userId, Integer score) {
        MaJiangGameItemDO item = new MaJiangGameItemDO();
        item.setGameId(gameId);
        item.setUserId(userId);
        item.setPoints(score);
        
        LocalDateTime now = LocalDateTime.now();
        item.setCreatedTime(now);
        item.setUpdatedTime(now);
        
        return item;
    }
}