package datalens.controller;

import datalens.config.SecurityConfig;
import datalens.model.DatasetDTO;
import datalens.service.DatasetService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.Import;
import org.springframework.mock.web.MockMultipartFile;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors;
import org.springframework.test.web.servlet.MockMvc;

import java.time.LocalDateTime;
import java.util.List;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(DatasetController.class)
@Import(SecurityConfig.class)
class DatasetControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private DatasetService datasetService;

    @MockBean
    private datalens.security.JwtTokenProvider jwtTokenProvider;

    @MockBean
    private datalens.security.CustomUserDetailsService customUserDetailsService;

    private DatasetDTO sampleDTO() {
        return DatasetDTO.builder()
                .id(1L)
                .name("My Dataset")
                .description("desc")
                .filename("data.csv")
                .mimeType("text/csv")
                .rowCount(100L)
                .sizeBytes(2048L)
                .createdAt(LocalDateTime.of(2025, 1, 1, 0, 0))
                .updatedAt(LocalDateTime.of(2025, 1, 1, 0, 0))
                .build();
    }

    @Test
    void listDatasetsReturnsList() throws Exception {
        when(datasetService.getUserDatasets("user"))
                .thenReturn(List.of(sampleDTO()));

        mockMvc.perform(get("/api/datasets")
                        .with(SecurityMockMvcRequestPostProcessors.user("user")
                                .authorities(new SimpleGrantedAuthority("ROLE_USER"))))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$[0].name").value("My Dataset"))
                .andExpect(jsonPath("$[0].filename").value("data.csv"));
    }

    @Test
    void uploadReturnsDTO() throws Exception {
        when(datasetService.uploadDataset(eq("user"), eq("Test"), isNull(), any()))
                .thenReturn(sampleDTO());

        MockMultipartFile file = new MockMultipartFile(
                "file", "data.csv", "text/csv", "col\n1".getBytes());

        mockMvc.perform(multipart("/api/datasets/upload")
                        .file(file)
                        .param("name", "Test")
                        .with(SecurityMockMvcRequestPostProcessors.user("user")
                                .authorities(new SimpleGrantedAuthority("ROLE_USER")))
                        .with(SecurityMockMvcRequestPostProcessors.csrf()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.name").value("My Dataset"));
    }

    @Test
    void getDatasetReturnsDTO() throws Exception {
        when(datasetService.getDataset("user", 1L)).thenReturn(sampleDTO());

        mockMvc.perform(get("/api/datasets/1")
                        .with(SecurityMockMvcRequestPostProcessors.user("user")
                                .authorities(new SimpleGrantedAuthority("ROLE_USER"))))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.name").value("My Dataset"))
                .andExpect(jsonPath("$.id").value(1));
    }

    @Test
    void deleteDatasetReturns204() throws Exception {
        doNothing().when(datasetService).deleteDataset("user", 1L);

        mockMvc.perform(delete("/api/datasets/1")
                        .with(SecurityMockMvcRequestPostProcessors.user("user")
                                .authorities(new SimpleGrantedAuthority("ROLE_USER")))
                        .with(SecurityMockMvcRequestPostProcessors.csrf()))
                .andExpect(status().isNoContent());
    }

    @Test
    void listDatasetsWithoutAuthReturnsForbidden() throws Exception {
        mockMvc.perform(get("/api/datasets"))
                .andExpect(status().isForbidden());
    }

    @Test
    void getDatasetWithoutAuthReturnsForbidden() throws Exception {
        mockMvc.perform(get("/api/datasets/1"))
                .andExpect(status().isForbidden());
    }

    @Test
    void deleteDatasetWithoutAuthReturnsForbidden() throws Exception {
        mockMvc.perform(delete("/api/datasets/1"))
                .andExpect(status().isForbidden());
    }
}
