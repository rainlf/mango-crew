package com.rainlf.mgttbe;

import com.rainlf.mgttbe.controller.dto.ApiResponse;
import com.rainlf.mgttbe.controller.dto.MaJiangGameLogDTO;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.*;

import java.util.List;

/**
 * 分页功能测试类
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
public class PaginationTest {

    @Autowired
    private TestRestTemplate restTemplate;

    @Test
    public void testPagination() {
        System.out.println("开始测试分页功能...");
        
        // 测试第一页（10条记录）
        System.out.println("测试第一页数据（limit=10, offset=0）");
        testPage(10, 0, 10);
        
        // 测试第二页（10条记录）
        System.out.println("\n测试第二页数据（limit=10, offset=10）");
        testPage(10, 10, 10);
        
        // 测试第三页（10条记录）
        System.out.println("\n测试第三页数据（limit=10, offset=20）");
        testPage(10, 20, 10);
        
        // 测试第四页（10条记录）
        System.out.println("\n测试第四页数据（limit=10, offset=30）");
        testPage(10, 30, 10);
        
        // 测试第十页（10条记录）
        System.out.println("\n测试第十页数据（limit=10, offset=90）");
        testPage(10, 90, 10);
        
        // 测试超出范围的分页（应该返回空列表）
        System.out.println("\n测试超出范围的分页（limit=10, offset=100）");
        testPage(10, 100, 0);
        
        // 测试自定义每页记录数
        System.out.println("\n测试自定义每页记录数（limit=5, offset=0）");
        testPage(5, 0, 5);
        
        System.out.println("\n分页功能测试完成！");
    }

    private void testPage(int limit, int offset, int expectedSize) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        HttpEntity<String> entity = new HttpEntity<>(headers);
        
        ResponseEntity<ApiResponse<List<MaJiangGameLogDTO>>> response = restTemplate.exchange(
                "/majiang/games?limit={limit}&offset={offset}",
                HttpMethod.GET,
                entity,
                new ParameterizedTypeReference<ApiResponse<List<MaJiangGameLogDTO>>>() {},
                limit, offset
        );
        
        if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null && response.getBody().isSuccess()) {
            List<MaJiangGameLogDTO> gameLogs = response.getBody().getData();
            System.out.println("成功获取分页数据，总数：" + gameLogs.size());
            System.out.println("预期数量：" + expectedSize);
            System.out.println("分页数据ID列表：" + gameLogs.stream().map(MaJiangGameLogDTO::getId).toList());
            
            // 验证返回的记录数量是否符合预期
            assert gameLogs.size() == expectedSize : 
                "分页数据数量不匹配：期望 " + expectedSize + "，实际 " + gameLogs.size();
        } else {
            System.out.println("获取分页数据失败，状态码：" + response.getStatusCode());
            assert false : "分页接口调用失败";
        }
    }
}