package com.rainlf.mgttbe.controller;

import com.rainlf.mgttbe.controller.dto.ApiResponse;
import com.rainlf.mgttbe.controller.dto.MaJiangGameLogDTO;
import com.rainlf.mgttbe.controller.dto.PlayersDTO;
import com.rainlf.mgttbe.controller.dto.SaveMaJiangGameRequest;
import com.rainlf.mgttbe.infra.aop.ExecutionTime;
import com.rainlf.mgttbe.infra.util.JsonUtils;
import com.rainlf.mgttbe.service.MaJiangService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;

@Slf4j
@RestController
@RequestMapping("/majiang")
public class MaJiangController {

    @Autowired
    private MaJiangService majiangService;

    @GetMapping("/games")
    @ExecutionTime
    public ApiResponse<List<MaJiangGameLogDTO>> getMaJiangGames(
            @RequestParam(value = "limit", required = false, defaultValue = "10") Integer limit,
            @RequestParam(value = "offset", required = false, defaultValue = "0") Integer offset) {
        List<MaJiangGameLogDTO> result = majiangService.getMaJiangGameLogs(limit, offset);
        // result.forEach(x -> {
        //     x.getPlayer1().setAvatar(null);
        //     x.getPlayer2().setAvatar(null);
        //     x.getPlayer3().setAvatar(null);
        //     x.getPlayer4().setAvatar(null);
        //     x.getWinners().forEach(y -> y.getUser().setAvatar(null));
        //     x.getLosers().forEach(y -> y.getUser().setAvatar(null));
        // });
        return ApiResponse.success(result);
    }

    @GetMapping("/user/games")
    @ExecutionTime
    public ApiResponse<List<MaJiangGameLogDTO>> getMaJiangGamesByUser
            (@RequestParam("userId") Integer userId,
             @RequestParam(value = "limit", required = false, defaultValue = "10") Integer limit,
             @RequestParam(value = "offset", required = false, defaultValue = "0") Integer offset) {
        List<MaJiangGameLogDTO> result = majiangService.getMaJiangGamesByUser(userId, limit, offset);
        // result.forEach(x -> {
        //     x.getPlayer1().setAvatar(null);
        //     x.getPlayer2().setAvatar(null);
        //     x.getPlayer3().setAvatar(null);
        //     x.getPlayer4().setAvatar(null);
        //     x.getWinners().forEach(y -> y.getUser().setAvatar(null));
        //     x.getLosers().forEach(y -> y.getUser().setAvatar(null));
        // });
        return ApiResponse.success(result);
    }

    @PostMapping("/game")
    public ApiResponse<Integer> saveMaJiangGame(@RequestBody SaveMaJiangGameRequest request) {
        log.info("saveMaJiangGame, request: {}", JsonUtils.writeString(request));
        validSaveMaJiangGameRequest(request);
        return ApiResponse.success(majiangService.saveMaJiangGame(request));
    }

    @DeleteMapping("/game")
    public ApiResponse<Void> deleteMaJiangGame(@RequestParam("id") Integer id, @RequestParam("userId") Integer userId) {
        log.info("deleteMaJiangGame, id: {}, userId: {}", id, userId);
        majiangService.deleteMaJiangGame(id, userId);
        return ApiResponse.success();
    }

    @GetMapping("/game/players")
    public ApiResponse<PlayersDTO> getGamePlayers() {
        return ApiResponse.success(majiangService.getMaJiangGamePlayers());
    }


    private void validSaveMaJiangGameRequest(SaveMaJiangGameRequest request) {
        // 运动类型(6)特殊处理：只需要一个玩家，losers可以为空
        boolean isSportType = request.getGameType() != null && request.getGameType() == 6;
        
        if (request.getPlayers() == null || 
            (isSportType && request.getPlayers().size() != 1) || 
            (!isSportType && request.getPlayers().size() != 4)) {
            throw new RuntimeException("Invalid number of players");
        }

        if (request.getRecorderId() == null) {
            throw new RuntimeException("Invalid recorder id");
        }

        if (request.getWinners() == null || request.getWinners().isEmpty()) {
            throw new RuntimeException("Invalid winners, winners is empty");
        }

        // 运动类型允许losers为空
        if (!isSportType && (request.getLosers() == null || request.getLosers().isEmpty())) {
            throw new RuntimeException("Invalid losers, losers is empty");
        }

        if (request.getWinners().stream().anyMatch(x -> x.getUserId() == null)) {
            throw new RuntimeException("Invalid winners, winners must contain at least one user");
        }

        if (request.getWinners().stream().anyMatch(x -> x.getBasePoints() <= 0)) {
            throw new RuntimeException("Invalid winners, winners basepoint must be greater than 0");
        }

        List<Integer> players = request.getPlayers();
        List<Integer> winners = request.getWinners().stream().map(SaveMaJiangGameRequest.Winner::getUserId).toList();
        List<Integer> losers = request.getLosers();

        // 赢家必须是玩家之一
        if (winners.stream().anyMatch(x -> !players.contains(x))) {
            throw new RuntimeException("Invalid winners, must be one of players");
        }

        // 运动类型且losers为空时，不检查losers相关验证
        if (!isSportType || (losers != null && !losers.isEmpty())) {
            // 检查winners和losers是否有交集
            if (!Collections.disjoint(winners, losers)) {
                throw new RuntimeException("Invalid winners or losers, cannot have common user");
            }
            // 检查losers是否都是玩家之一
            if (losers.stream().anyMatch(y -> !players.contains(y))) {
                throw new RuntimeException("Invalid losers, must be one of players");
            }
        }
    }
}


