package com.rainlf.mgttbe.infra.db;

import com.rainlf.mgttbe.infra.db.dataobj.UserDO;
import com.rainlf.mgttbe.infra.db.manager.MaJiangGameItemManager;
import com.rainlf.mgttbe.infra.db.manager.MaJiangGameManager;
import com.rainlf.mgttbe.infra.db.repository.UserRepository;
import com.rainlf.mgttbe.infra.util.JsonUtils;
import com.rainlf.mgttbe.model.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

/**
 * 数据初始化器，用于在应用启动时创建测试数据
 */
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;

@Component
@ConditionalOnProperty(name = "mgtt.data.initialize", havingValue = "true")
public class DataInitializer implements ApplicationRunner {

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private MaJiangGameManager maJiangGameManager;

    @Autowired
    private MaJiangGameItemManager maJiangGameItemManager;

    @Override
    public void run(ApplicationArguments args) throws Exception {
        // 检查是否已有数据，如果有则不再初始化
        if (!userRepository.findAll().isEmpty()) {
            return;
        }

        // 创建用户
        createTestUsers();
        // 创建游戏记录
        createTestGames();
    }

    private void createTestUsers() {
        List<UserDO> users = new ArrayList<>();
        
        // 创建8个测试用户，每个用户有不同的用户名和初始积分
        for (int i = 1; i <= 8; i++) {
            UserDO user = new UserDO();
            user.setUsername("玩家" + i);
            user.setAvatar(null); // avatar字段是byte[]类型，这里设为null
            user.setPoints(1000); // 所有用户初始积分为1000
            user.setIsDeleted(0);
            users.add(user);
        }

        userRepository.saveAll(users);
    }

    private void createTestGames() {
        List<UserDO> users = userRepository.findAll();
        if (users.size() < 4) {
            return;
        }

        // 创建100条游戏记录，用于测试分页功能
        for (int i = 0; i < 100; i++) {
            // 从所有用户中随机选择4个不同的用户参与游戏
            List<Integer> randomUserIds = selectRandomUsers(users, 4);
            createGameRecord(randomUserIds.get(0), randomUserIds.get(1), randomUserIds.get(2), randomUserIds.get(3), i);
        }
    }
    
    /**
     * 从用户列表中随机选择指定数量的不同用户ID
     */
    private List<Integer> selectRandomUsers(List<UserDO> users, int count) {
        List<Integer> allUserIds = users.stream().map(UserDO::getId).collect(Collectors.toList());
        Collections.shuffle(allUserIds);
        return allUserIds.subList(0, count);
    }

    private void createGameRecord(Integer userId1, Integer userId2, Integer userId3, Integer userId4, int index) {
        MaJiangGame game = new MaJiangGame();
        // 随机选择游戏类型
        MaJiangGameType[] gameTypes = {MaJiangGameType.PING_HU, MaJiangGameType.ZI_MO, MaJiangGameType.YI_PAO_SHUANG_XIANG};
        game.setType(gameTypes[index % gameTypes.length]);
        game.setPlayer1(userId1);
        game.setPlayer2(userId2);
        game.setPlayer3(userId3);
        game.setPlayer4(userId4);
        game.setDeleted(false);
        // 设置创建时间和更新时间，确保不为null
        LocalDateTime now = LocalDateTime.now();
        // 为了测试分页，创建不同的时间点
        LocalDateTime createdTime = now.minusDays(index);
        game.setCreatedTime(createdTime);
        game.setUpdatedTime(createdTime);
        
        // 保存游戏记录
        game = maJiangGameManager.save(game);

        // 创建游戏项目记录
        List<MaJiangGameItem> items = new ArrayList<>();

        // 随机选择赢家
        List<Integer> playerIds = Arrays.asList(userId1, userId2, userId3, userId4);
        Collections.shuffle(playerIds);

        // 根据游戏类型创建不同的赢家和输家
        if (game.getType() == MaJiangGameType.PING_HU) {
            // 平胡：1赢1输
            MaJiangGameItem winner = createGameItem(game.getId(), playerIds.get(0), MaJiangUserType.WINNER, 20);
            items.add(winner);
            MaJiangGameItem loser = createGameItem(game.getId(), playerIds.get(1), MaJiangUserType.LOSER, -20);
            items.add(loser);
        } else if (game.getType() == MaJiangGameType.ZI_MO) {
            // 自摸：1赢3输
            MaJiangGameItem winner = createGameItem(game.getId(), playerIds.get(0), MaJiangUserType.WINNER, 30);
            items.add(winner);
            for (int i = 1; i <= 3; i++) {
                MaJiangGameItem loser = createGameItem(game.getId(), playerIds.get(i), MaJiangUserType.LOSER, -10);
                items.add(loser);
            }
        } else if (game.getType() == MaJiangGameType.YI_PAO_SHUANG_XIANG) {
            // 一炮双响：2赢1输
            MaJiangGameItem winner1 = createGameItem(game.getId(), playerIds.get(0), MaJiangUserType.WINNER, 15);
            items.add(winner1);
            MaJiangGameItem winner2 = createGameItem(game.getId(), playerIds.get(1), MaJiangUserType.WINNER, 15);
            items.add(winner2);
            MaJiangGameItem loser = createGameItem(game.getId(), playerIds.get(2), MaJiangUserType.LOSER, -30);
            items.add(loser);
        }

        // 创建记录员（随机选择一个玩家）
        MaJiangGameItem recorder = createGameItem(game.getId(), playerIds.get(new Random().nextInt(playerIds.size())), MaJiangUserType.RECORDER, 1);
        items.add(recorder);

        // 保存游戏项目记录
        maJiangGameItemManager.save(items);
    }

    private MaJiangGameItem createGameItem(Integer gameId, Integer userId, MaJiangUserType type, Integer points) {
        MaJiangGameItem item = new MaJiangGameItem();
        item.setGameId(gameId);
        item.setUserId(userId);
        item.setType(type);
        item.setPoints(points);
        if (type == MaJiangUserType.WINNER) {
            item.setBasePoint(points);
            // 随机添加一些赢法类型
            List<MaJiangWinType> winTypes = new ArrayList<>();
            winTypes.add(MaJiangWinType.WU_HUA_GUO); // 无花果
            if (new Random().nextBoolean()) {
                winTypes.add(MaJiangWinType.PENG_PENG_HU); // 碰碰胡
            }
            // 直接设置赢法类型列表
            item.setWinTypes(winTypes);
        }
        return item;
    }
}