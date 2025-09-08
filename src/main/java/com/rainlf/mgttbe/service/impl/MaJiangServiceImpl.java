package com.rainlf.mgttbe.service.impl;

import com.rainlf.mgttbe.controller.dto.MaJiangGameLogDTO;
import com.rainlf.mgttbe.controller.dto.PlayersDTO;
import com.rainlf.mgttbe.controller.dto.SaveMaJiangGameRequest;
import com.rainlf.mgttbe.controller.dto.UserDTO;
import com.rainlf.mgttbe.infra.aop.ExecutionTime;
import com.rainlf.mgttbe.infra.db.manager.MaJiangGameItemManager;
import com.rainlf.mgttbe.infra.db.manager.MaJiangGameManager;
import com.rainlf.mgttbe.infra.util.DateUtils;
import com.rainlf.mgttbe.model.*;
import com.rainlf.mgttbe.service.MaJiangService;
import com.rainlf.mgttbe.service.UserService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class MaJiangServiceImpl implements MaJiangService {
    @Autowired
    private MaJiangGameManager majiangGameManager;
    @Autowired
    private MaJiangGameItemManager majiangGameItemManager;
    @Autowired
    private UserService userService;
    private final Random random = new Random();

    @Override
    @Transactional(rollbackFor = Exception.class)
    public Integer saveMaJiangGame(SaveMaJiangGameRequest request) {
        MaJiangGame majiangGame = new MaJiangGame();
        majiangGame.setType(MaJiangGameType.fromCode(request.getGameType()));
        
        // 对于运动类型，只有player1设置为记录者，其他三个玩家用-1代替
        if (majiangGame.getType() == MaJiangGameType.YUN_DONG) {
            Integer recorderId = request.getRecorderId();
            majiangGame.setPlayer1(recorderId);
            majiangGame.setPlayer2(-1);
            majiangGame.setPlayer3(-1);
            majiangGame.setPlayer4(-1);
        } else {
            majiangGame.setPlayer1(request.getPlayers().get(0));
            majiangGame.setPlayer2(request.getPlayers().get(1));
            majiangGame.setPlayer3(request.getPlayers().get(2));
            majiangGame.setPlayer4(request.getPlayers().get(3));
        }
        
        majiangGame = majiangGameManager.save(majiangGame);

        Integer gameId = majiangGame.getId();

        // winner
        List<MaJiangGameItem> winners = new ArrayList<>();
        for (SaveMaJiangGameRequest.Winner winner : request.getWinners()) {
            MaJiangGameItem maJiangGameItem = new MaJiangGameItem();
            maJiangGameItem.setGameId(gameId);
            maJiangGameItem.setUserId(winner.getUserId());
            maJiangGameItem.setType(MaJiangUserType.WINNER);
            maJiangGameItem.setBasePoint(winner.getBasePoints());
            maJiangGameItem.setWinTypes(winner.getWinTypes().stream().map(MaJiangWinType::fromName).toList());
            winners.add(maJiangGameItem);
        }

        // loser
        List<MaJiangGameItem> losers = new ArrayList<>();
        // 对于运动类型，不创建输家
        if (majiangGame.getType() != MaJiangGameType.YUN_DONG) {
            for (Integer loser : request.getLosers()) {
                // 过滤掉无效的用户ID
                if (loser != null && loser > 0) {
                    MaJiangGameItem maJiangGameItem = new MaJiangGameItem();
                    maJiangGameItem.setGameId(gameId);
                    maJiangGameItem.setUserId(loser);
                    maJiangGameItem.setType(MaJiangUserType.LOSER);
                    losers.add(maJiangGameItem);
                }
            }
        }

        switch (majiangGame.getType()) {
            case PING_HU -> {
                if (winners.size() != 1 || losers.size() != 1) {
                    throw new RuntimeException("Ping HU losers and losers don't match");
                }
                MaJiangGameItem winner = winners.getFirst();
                winner.calculatePoints();
                losers.getFirst().setPoints(-winner.getPoints());
            }
            case ZI_MO -> {
                if (winners.size() != 1 || losers.size() != 3) {
                    throw new RuntimeException("ZI MO losers and losers don't match");
                }
                MaJiangGameItem winner = winners.getFirst();
                winner.calculatePoints();
                int points = winner.getPoints();
                winner.setPoints(points * 3);
                losers.forEach(x -> x.setPoints(-points));
            }
            case YI_PAO_SHUANG_XIANG -> {
                if (winners.size() != 2 || losers.size() != 1) {
                    throw new RuntimeException("YI PAO shuang xiang losers don't match");
                }
                int points = 0;
                for (MaJiangGameItem x : winners) {
                    x.calculatePoints();
                    points += x.getPoints();
                }
                losers.getFirst().setPoints(-points);
            }
            case YI_PAO_SAN_XIANG -> {
                if (winners.size() != 3 || losers.size() != 1) {
                    throw new RuntimeException("YI PAO san xiang losers don't match");
                }
                int points = 0;
                for (MaJiangGameItem x : winners) {
                    x.calculatePoints();
                    points += x.getPoints();
                }
                losers.getFirst().setPoints(-points);
            }
            case XIANG_GONG -> {
                if (winners.size() != 3 || losers.size() != 1) {
                    throw new RuntimeException("Xiang Gong losers don't match");
                }
                int points = 0;
                for (MaJiangGameItem x : winners) {
                    x.calculatePoints();
                    points += x.getPoints();
                }
                losers.getFirst().setPoints(-points);
            }
            case YUN_DONG -> {
                // 运动类型的特殊逻辑：没有输家，用户自行填入分数
                if (winners.size() != 1) {
                    throw new RuntimeException("运动类型只能记录一个用户");
                }
                // 验证只能记录自己
                Integer recorderId = request.getRecorderId();
                if (!Objects.equals(winners.getFirst().getUserId(), recorderId)) {
                    throw new RuntimeException("运动类型只能记录自己");
                }
                // 使用用户填入的分数，不进行计算
                MaJiangGameItem winner = winners.getFirst();
                winner.setPoints(winner.getBasePoint());
            }
        }

        List<MaJiangGameItem> items = new ArrayList<>();
        items.addAll(winners);
        items.addAll(losers);
        
        if (majiangGame.getType() != MaJiangGameType.YUN_DONG) {
            // recorder
            MaJiangGameItem recorder = new MaJiangGameItem();
            recorder.setGameId(gameId);
            recorder.setUserId(request.getRecorderId());
            recorder.setType(MaJiangUserType.RECORDER);
            // 生成一个 0 到 99 之间的随机整数, 1% 加 20, 99% 加 1
            int recorderPoints = (random.nextInt(100) < 1) ? 20 : 1;
            recorder.setPoints(recorderPoints);
            items.add(recorder);
        } else {
            // 运动类型有recorder，但不添加分数
            MaJiangGameItem recorder = new MaJiangGameItem();
            recorder.setGameId(gameId);
            recorder.setUserId(request.getRecorderId());
            recorder.setType(MaJiangUserType.RECORDER);
            recorder.setPoints(0); // 分数设置为0，不影响用户积分
            items.add(recorder);
        }
        
        majiangGameItemManager.save(items);

        // update user points with atom operation
        for (MaJiangGameItem item : items) {
            User user = userService.findUserById(item.getUserId());
            log.info("game save, id: {}, user: {}, points: {}", gameId, user.getUsername(), item.getPoints());
            userService.updateUserPoint(item.getUserId(), item.getPoints());
        }

        return gameId;
    }

    @Override
    @ExecutionTime
    public List<MaJiangGameLogDTO> getMaJiangGameLogs(Integer limit, Integer offset) {
        List<MaJiangGame> games = majiangGameManager.findLastGames(limit, offset);
        return buildMaJiangGameLogDTO(games);
    }

    @Override
    @ExecutionTime
    public List<MaJiangGameLogDTO> getMaJiangGamesByUser(Integer userId, Integer limit, Integer offset) {
        List<Integer> lastGameIds = majiangGameItemManager.findLastGameIdsByUserIdAndTypeIn(userId, List.of(MaJiangUserType.WINNER, MaJiangUserType.LOSER), limit,  offset);
        List<MaJiangGame> games = majiangGameManager.findByIdIn(lastGameIds);
        return buildMaJiangGameLogDTO(userId, games);
    }

    @ExecutionTime
    private List<MaJiangGameLogDTO> buildMaJiangGameLogDTO(Integer userId, List<MaJiangGame> gamesInput) {
        List<MaJiangGame> games = new ArrayList<>(gamesInput);
        games.sort(Comparator.comparing(MaJiangGame::getCreatedTime, Comparator.nullsLast(Comparator.reverseOrder())));
        List<MaJiangGameLogDTO> result = buildMaJiangGameLogDTO(games);
        for (MaJiangGameLogDTO maJiangGameLogDTO : result) {
            maJiangGameLogDTO.setForOnePlayer(true);
            if (maJiangGameLogDTO.getWinners().stream().map(x->x.getUser().getId()).collect(Collectors.toSet()).contains(userId)) {
                maJiangGameLogDTO.setPlayerWin(true);
            } else if (maJiangGameLogDTO.getLosers().stream().map(x->x.getUser().getId()).collect(Collectors.toSet()).contains(userId)) {
                maJiangGameLogDTO.setPlayerWin(false);
            } else {
                log.error("should not happen");
                maJiangGameLogDTO.setPlayerWin(false);
            }
        }
        return result;
    }

    @ExecutionTime
    private List<MaJiangGameLogDTO> buildMaJiangGameLogDTO(List<MaJiangGame> games) {
        List<MaJiangGameLogDTO> dtos = new ArrayList<>();
        for (MaJiangGame game : games) {
            MaJiangGameLogDTO dto = new MaJiangGameLogDTO();
            dto.setId(game.getId());
            if (Objects.equals(game.getType(), MaJiangGameType.ZI_MO)) {
                dto.setType(MaJiangGameType.ZI_MO.getName());
            } else {
                dto.setType(game.getType().getName());
            }

            dto.setPlayer1(UserDTO.fromUser(userService.findUserById(game.getPlayer1())));
            dto.setPlayer2(game.getPlayer2() != null && game.getPlayer2() > 0 ? UserDTO.fromUser(userService.findUserById(game.getPlayer2())) : null);
            dto.setPlayer3(game.getPlayer3() != null && game.getPlayer3() > 0 ? UserDTO.fromUser(userService.findUserById(game.getPlayer3())) : null);
            dto.setPlayer4(game.getPlayer4() != null && game.getPlayer4() > 0 ? UserDTO.fromUser(userService.findUserById(game.getPlayer4())) : null);
            dto.setCreatedTime(DateUtils.formatDateTime(game.getCreatedTime()));
            dto.setUpdatedTime(DateUtils.formatDateTime(game.getUpdatedTime()));

            List<MaJiangGameItem> gameItems = majiangGameItemManager.findByGameId(game.getId());
            List<MaJiangGameItem> winnerGameItems = gameItems.stream().filter(x -> Objects.equals(x.getType(), MaJiangUserType.WINNER)).toList();
            List<MaJiangGameItem> loserGameItems = gameItems.stream().filter(x -> Objects.equals(x.getType(), MaJiangUserType.LOSER)).toList();
            MaJiangGameItem recorderGameItems = gameItems.stream().filter(x -> Objects.equals(x.getType(), MaJiangUserType.RECORDER)).toList().getFirst();

            List<MaJiangGameLogDTO.Item> winnerItems = new ArrayList<>();
            for (MaJiangGameItem winnerGameItem : winnerGameItems) {
                MaJiangGameLogDTO.Item item = new MaJiangGameLogDTO.Item();
                item.setUser(UserDTO.fromUser(userService.findUserById(winnerGameItem.getUserId())));
                item.setPoints(winnerGameItem.getPoints());
                List<String> tags = new ArrayList<>();
                if (Objects.equals(game.getType(), MaJiangGameType.ZI_MO)) {
                    tags.add(game.getType().getName());
                } else {
                    tags.add(winnerGameItem.getType().getName());
                }
                if (winnerGameItem.getWinTypes() != null) {
                    tags.addAll(winnerGameItem.getWinTypes().stream().map(MaJiangWinType::getName).toList());
                }
                item.setTags(tags);
                winnerItems.add(item);
            }
            dto.setWinners(winnerItems);

            List<MaJiangGameLogDTO.Item> loserItems = new ArrayList<>();
            for (MaJiangGameItem loserGameItem : loserGameItems) {
                MaJiangGameLogDTO.Item item = new MaJiangGameLogDTO.Item();
                item.setUser(UserDTO.fromUser(userService.findUserById(loserGameItem.getUserId())));
                item.setPoints(loserGameItem.getPoints());
                List<String> tags = new ArrayList<>();
                if (Objects.equals(game.getType(), MaJiangGameType.YI_PAO_SHUANG_XIANG) || Objects.equals(game.getType(), MaJiangGameType.YI_PAO_SAN_XIANG)) {
                    tags.add(game.getType().getName());
                } else {
                    tags.add(loserGameItem.getType().getName());
                }
                item.setTags(tags);
                loserItems.add(item);
            }
            dto.setLosers(loserItems);

            MaJiangGameLogDTO.Item recorderItem = new MaJiangGameLogDTO.Item();
            recorderItem.setUser(UserDTO.fromUser(userService.findUserById(recorderGameItems.getUserId())));
            recorderItem.setPoints(recorderGameItems.getPoints());
            dto.setRecorder(recorderItem);
            dtos.add(dto);
        }
        return dtos;
    }

    @Override
    @Transactional
    public void deleteMaJiangGame(Integer id, Integer userId) {
        MaJiangGame maJiangGame = majiangGameManager.findByIdWithLock(id);
        if (maJiangGame == null) {
            throw new RuntimeException("majianggame not exist, game id: " + id);
        }
        List<MaJiangGameItem> items = majiangGameItemManager.findByGameIdAndType(id, MaJiangUserType.RECORDER);
        if (items.size() != 1) {
            throw new RuntimeException("majianggame recorder is null or multi, game id: " + id);
        }
        MaJiangGameItem recorder = items.getFirst();
        if (!Objects.equals(recorder.getUserId(), userId)) {
            throw new RuntimeException("only recorder can delete this game record");
        }

        List<MaJiangGameItem> maJiangGameItems = majiangGameItemManager.findByGameId(id);
        for (MaJiangGameItem maJiangGameItem : maJiangGameItems) {
            User user = userService.findUserById(maJiangGameItem.getUserId());
            log.info("game rollback, id: {}, user: {}, points: {}", id, user.getUsername(), -maJiangGameItem.getPoints());
            userService.updateUserPoint(maJiangGameItem.getUserId(), -maJiangGameItem.getPoints());
        }

        maJiangGame.setDeleted(true);
        majiangGameManager.save(maJiangGame);
    }

    @Override
    public PlayersDTO getMaJiangGamePlayers() {
        PlayersDTO playersDTO = new PlayersDTO();

        // 获取最近的一条非运动类型的记录
        List<MaJiangGame> maJiangGames = majiangGameManager.findLastGames(10, 0); // 获取最近10条记录以确保有非运动记录
        MaJiangGame validGame = null;
        for (MaJiangGame game : maJiangGames) {
            if (game.getType() != MaJiangGameType.YUN_DONG) {
                validGame = game;
                break;
            }
        }
        
        if (validGame != null) {
            List<UserDTO> currentPlayers = new ArrayList<>();
            currentPlayers.add(UserDTO.fromUser(userService.findUserById(validGame.getPlayer1())));
            currentPlayers.add(UserDTO.fromUser(userService.findUserById(validGame.getPlayer2())));
            currentPlayers.add(UserDTO.fromUser(userService.findUserById(validGame.getPlayer3())));
            currentPlayers.add(UserDTO.fromUser(userService.findUserById(validGame.getPlayer4())));
            playersDTO.setCurrentPlayers(currentPlayers);
        } else {
            // 如果没有找到非运动记录，则返回空列表
            playersDTO.setCurrentPlayers(new ArrayList<>());
        }

        List<User> users = userService.findAllUsers();
        playersDTO.setAllPlayers(users.stream().map(UserDTO::fromUser).collect(Collectors.toList()));
        return playersDTO;
    }
}
