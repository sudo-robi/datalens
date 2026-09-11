package datalens.service;

import datalens.model.Dataset;
import datalens.model.DatasetDTO;
import datalens.model.User;
import datalens.repository.DatasetRepository;
import datalens.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class DatasetService {

    private final DatasetRepository datasetRepository;
    private final UserRepository userRepository;

    private static final String UPLOAD_DIR = "uploads/";

    public DatasetDTO uploadDataset(String username, String name, String description,
                                     MultipartFile file) throws IOException {
        User user = userRepository.findByUsername(username)
                .orElseThrow(() -> new RuntimeException("User not found"));

        Path uploadPath = Paths.get(UPLOAD_DIR);
        if (!Files.exists(uploadPath)) {
            Files.createDirectories(uploadPath);
        }

        String filename = System.currentTimeMillis() + "_" + file.getOriginalFilename();
        Path filePath = uploadPath.resolve(filename);
        file.transferTo(filePath.toFile());

        Dataset dataset = Dataset.builder()
                .name(name)
                .description(description)
                .filename(file.getOriginalFilename())
                .storagePath(filePath.toString())
                .mimeType(file.getContentType())
                .sizeBytes(file.getSize())
                .user(user)
                .build();

        dataset = datasetRepository.save(dataset);
        return DatasetDTO.fromEntity(dataset);
    }

    public List<DatasetDTO> getUserDatasets(String username) {
        User user = userRepository.findByUsername(username)
                .orElseThrow(() -> new RuntimeException("User not found"));

        return datasetRepository.findByUserIdOrderByCreatedAtDesc(user.getId())
                .stream()
                .map(DatasetDTO::fromEntity)
                .collect(Collectors.toList());
    }

    public DatasetDTO getDataset(String username, Long datasetId) {
        Dataset dataset = datasetRepository.findById(datasetId)
                .orElseThrow(() -> new RuntimeException("Dataset not found"));

        if (!dataset.getUser().getUsername().equals(username)) {
            throw new RuntimeException("Access denied");
        }

        return DatasetDTO.fromEntity(dataset);
    }

    public void deleteDataset(String username, Long datasetId) {
        Dataset dataset = datasetRepository.findById(datasetId)
                .orElseThrow(() -> new RuntimeException("Dataset not found"));

        if (!dataset.getUser().getUsername().equals(username)) {
            throw new RuntimeException("Access denied");
        }

        try {
            Files.deleteIfExists(Paths.get(dataset.getStoragePath()));
        } catch (IOException e) {
            // Log but don't fail
        }

        datasetRepository.delete(dataset);
    }
}
