package datalens.controller;

import datalens.model.DatasetDTO;
import datalens.service.DatasetService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.util.List;

@RestController
@RequestMapping("/api/datasets")
@RequiredArgsConstructor
public class DatasetController {

    private final DatasetService datasetService;

    @PostMapping("/upload")
    public ResponseEntity<DatasetDTO> upload(
            @AuthenticationPrincipal UserDetails userDetails,
            @RequestParam("name") String name,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam("file") MultipartFile file) throws IOException {
        return ResponseEntity.ok(datasetService.uploadDataset(
                userDetails.getUsername(), name, description, file));
    }

    @GetMapping
    public ResponseEntity<List<DatasetDTO>> listDatasets(
            @AuthenticationPrincipal UserDetails userDetails) {
        return ResponseEntity.ok(datasetService.getUserDatasets(userDetails.getUsername()));
    }

    @GetMapping("/{id}")
    public ResponseEntity<DatasetDTO> getDataset(
            @AuthenticationPrincipal UserDetails userDetails,
            @PathVariable Long id) {
        return ResponseEntity.ok(datasetService.getDataset(userDetails.getUsername(), id));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteDataset(
            @AuthenticationPrincipal UserDetails userDetails,
            @PathVariable Long id) {
        datasetService.deleteDataset(userDetails.getUsername(), id);
        return ResponseEntity.noContent().build();
    }
}
