package datalens.service;

import datalens.model.Dataset;
import datalens.model.DatasetDTO;
import datalens.model.User;
import datalens.repository.DatasetRepository;
import datalens.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.mock.web.MockMultipartFile;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class DatasetServiceTest {

    @Mock
    private DatasetRepository datasetRepository;

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private DatasetService datasetService;

    private User testUser;
    private Dataset testDataset;

    @BeforeEach
    void setUp() {
        testUser = User.builder()
                .id(1L)
                .username("testuser")
                .email("test@example.com")
                .passwordHash("hashed")
                .build();

        testDataset = Dataset.builder()
                .id(10L)
                .name("Test Dataset")
                .description("A test dataset")
                .filename("data.csv")
                .storagePath("uploads/test.csv")
                .mimeType("text/csv")
                .sizeBytes(1024L)
                .user(testUser)
                .rowCount(100L)
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }

    @Test
    void uploadDatasetSavesFileAndReturnsDTO() throws IOException {
        when(userRepository.findByUsername("testuser")).thenReturn(Optional.of(testUser));
        when(datasetRepository.save(any(Dataset.class))).thenAnswer(inv -> {
            Dataset d = inv.getArgument(0);
            d.setId(20L);
            return d;
        });

        MockMultipartFile file = new MockMultipartFile(
                "file", "data.csv", "text/csv", "col1,col2\n1,2".getBytes());

        DatasetDTO result = datasetService.uploadDataset("testuser", "New Dataset", "desc", file);

        assertNotNull(result);
        assertEquals("New Dataset", result.getName());
        assertEquals("desc", result.getDescription());
        assertEquals("data.csv", result.getFilename());
        verify(datasetRepository).save(any(Dataset.class));
    }

    @Test
    void uploadDatasetThrowsWhenUserNotFound() {
        when(userRepository.findByUsername("missing")).thenReturn(Optional.empty());

        MockMultipartFile file = new MockMultipartFile(
                "file", "data.csv", "text/csv", "data".getBytes());

        assertThrows(RuntimeException.class, () ->
                datasetService.uploadDataset("missing", "name", "desc", file));
    }

    @Test
    void getUserDatasetsReturnsListOfDTOs() {
        when(userRepository.findByUsername("testuser")).thenReturn(Optional.of(testUser));
        when(datasetRepository.findByUserIdOrderByCreatedAtDesc(1L))
                .thenReturn(List.of(testDataset));

        List<DatasetDTO> result = datasetService.getUserDatasets("testuser");

        assertEquals(1, result.size());
        assertEquals("Test Dataset", result.get(0).getName());
    }

    @Test
    void getUserDatasetsThrowsWhenUserNotFound() {
        when(userRepository.findByUsername("missing")).thenReturn(Optional.empty());

        assertThrows(RuntimeException.class, () ->
                datasetService.getUserDatasets("missing"));
    }

    @Test
    void getDatasetReturnsDTOForOwner() {
        when(datasetRepository.findById(10L)).thenReturn(Optional.of(testDataset));

        DatasetDTO result = datasetService.getDataset("testuser", 10L);

        assertNotNull(result);
        assertEquals("Test Dataset", result.getName());
        assertEquals("testuser", testDataset.getUser().getUsername());
    }

    @Test
    void getDatasetThrowsForNonOwner() {
        User otherUser = User.builder()
                .id(2L).username("other").email("o@x.com").passwordHash("h").build();
        Dataset otherDataset = Dataset.builder()
                .id(11L).name("Other").filename("o.csv").storagePath("o")
                .mimeType("text/csv").user(otherUser).build();

        when(datasetRepository.findById(11L)).thenReturn(Optional.of(otherDataset));

        assertThrows(RuntimeException.class, () ->
                datasetService.getDataset("testuser", 11L));
    }

    @Test
    void getDatasetThrowsWhenDatasetNotFound() {
        when(datasetRepository.findById(999L)).thenReturn(Optional.empty());

        assertThrows(RuntimeException.class, () ->
                datasetService.getDataset("testuser", 999L));
    }

    @Test
    void deleteDatasetDeletesForOwner() throws IOException {
        Path tempFile = Files.createTempFile("test-delete", ".csv");
        Dataset deletable = Dataset.builder()
                .id(30L).name("Del").filename("del.csv")
                .storagePath(tempFile.toString())
                .mimeType("text/csv").user(testUser).build();

        when(datasetRepository.findById(30L)).thenReturn(Optional.of(deletable));

        datasetService.deleteDataset("testuser", 30L);

        verify(datasetRepository).delete(deletable);
    }

    @Test
    void deleteDatasetThrowsForNonOwner() {
        User other = User.builder()
                .id(3L).username("other").email("o@x.com").passwordHash("h").build();
        Dataset otherDs = Dataset.builder()
                .id(31L).name("O").filename("o.csv").storagePath("/tmp/o")
                .mimeType("text/csv").user(other).build();

        when(datasetRepository.findById(31L)).thenReturn(Optional.of(otherDs));

        assertThrows(RuntimeException.class, () ->
                datasetService.deleteDataset("testuser", 31L));
    }

    @Test
    void deleteDatasetThrowsWhenDatasetNotFound() {
        when(datasetRepository.findById(999L)).thenReturn(Optional.empty());

        assertThrows(RuntimeException.class, () ->
                datasetService.deleteDataset("testuser", 999L));
    }

    @Test
    void updateRowCountUpdatesExistingDataset() {
        when(datasetRepository.findById(10L)).thenReturn(Optional.of(testDataset));

        datasetService.updateRowCount(10L, 500L);

        assertEquals(500L, testDataset.getRowCount());
        verify(datasetRepository).save(testDataset);
    }

    @Test
    void updateRowCountDoesNothingForNonExistentDataset() {
        when(datasetRepository.findById(999L)).thenReturn(Optional.empty());

        datasetService.updateRowCount(999L, 500L);

        verify(datasetRepository, never()).save(any());
    }
}
