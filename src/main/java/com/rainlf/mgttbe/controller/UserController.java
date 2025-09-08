package com.rainlf.mgttbe.controller;

import com.rainlf.mgttbe.controller.dto.ApiResponse;
import com.rainlf.mgttbe.controller.dto.UserDTO;
import com.rainlf.mgttbe.model.User;
import com.rainlf.mgttbe.service.UserService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.util.List;


@Slf4j
@RestController
@RequestMapping("/user")
public class UserController {

    @Autowired
    private UserService userService;

    @GetMapping("/login")
    public ApiResponse<UserDTO> login(@RequestParam("code") String code) {
        User user = userService.login(code);
        return ApiResponse.success(UserDTO.fromUser(user));
    }

    @PostMapping("/info")
    public UserDTO updateUserInfo(@RequestParam("avatar") MultipartFile avatar, @RequestParam("userId") Integer userId, @RequestParam("username") String username) throws IOException {
        log.info("updateUserInfo, avatar: {}, userId: {}, username: {}", avatar, userId, username);
        if (avatar == null) {
            throw new RuntimeException("avatar cant be null");
        }
        if (userId == null) {
            throw new RuntimeException("userId cant be null");
        }
        if (!StringUtils.hasText(username)) {
            throw new RuntimeException("username cant be null");
        }
        if (username.length() > 16) {
            throw new RuntimeException("username is too long");
        }
        User user = userService.save(userId, username, avatar.getBytes());
        return UserDTO.fromUser(user);
    }


    @PostMapping("/username")
    public ApiResponse<Void> getAvatar(@RequestParam("userId") Integer userId, @RequestParam("username") String username) {
        userService.save(userId, username);
        return ApiResponse.success();
    }

    @GetMapping("/info")
    public ApiResponse<UserDTO> getUserInfo(@RequestParam("userId") Integer userId) {
        User user = userService.findUserById(userId);
        return ApiResponse.success(UserDTO.fromUser(user));
    }

    @GetMapping("/rank")
    public ApiResponse<List<UserDTO>> getUserRank() {
        List<User> users = userService.getUserRank();
        return ApiResponse.success(users.stream().map(UserDTO::fromUser).toList());
    }
}
