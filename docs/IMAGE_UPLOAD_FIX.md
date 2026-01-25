# Image Upload Fix - Manual Question Creation

**Date**: January 24, 2026
**Version**: 1.3.2

## Problem

When manually creating questions with images through the frontend, the image was being uploaded successfully, but the question's `meta` field was not showing `image_url` or `image_file_id`.

**Example of broken response**:
```json
{
  "id": "fd367480-710a-4c67-84b7-cc5566c43dd8",
  "question_text": "adding image",
  "meta": {
    "answer_count": 2
  }
}
```

**Expected response**:
```json
{
  "id": "fd367480-710a-4c67-84b7-cc5566c43dd8",
  "question_text": "adding image",
  "meta": {
    "image_url": "/api/v1/files/serve/abc-123-def",
    "image_file_id": "abc-123-def",
    "answer_count": 2
  }
}
```

## Root Cause

The bug was in `backend/api/quiz.go` in both `AddQuestion` and `UpdateQuestion` handlers:

1. **Step 1** (lines 458-476): Backend receives multipart form data with an image file
   - Saves the image file to disk
   - Sets `req.ImageFileID = &imageFileID`

2. **Step 2** (lines 479-485): Backend parses `question_data` JSON field
   - Unmarshals JSON into `req` struct
   - **Bug**: This overwrites the ENTIRE struct, including the `ImageFileID` that was just set!

3. **Result**: The core layer never receives the `ImageFileID`, so it can't add image metadata to the question

## Solution

Fixed by preserving the uploaded image file ID before parsing JSON, then restoring it after:

```go
var uploadedImageFileID *string

// Upload image file
if imageFile != nil {
    fileUpload, err := fileUploadService.SaveFile(ctx, imageFile, imageHeader, userID.String())
    imageFileID := fileUpload.ID
    uploadedImageFileID = &imageFileID  // Save it!
}

// Parse question_data JSON
if questionDataJSON := c.PostForm("question_data"); questionDataJSON != "" {
    json.Unmarshal([]byte(questionDataJSON), &req)
    // Restore the uploaded image ID (takes precedence over any ID in JSON)
    if uploadedImageFileID != nil {
        req.ImageFileID = uploadedImageFileID
    }
}
```

## Files Modified

1. `backend/api/quiz.go`
   - Fixed `AddQuestion` handler (lines 456-491)
   - Fixed `UpdateQuestion` handler (lines 741-780)

2. `backend/CHANGELOG.md`
   - Added v1.3.2 release notes

## Testing

### Test 1: Add question with image via frontend

1. Navigate to edit quiz page
2. Click "Add Questions Manually"
3. Select a question type
4. Fill in question text and answers
5. Upload an image file
6. Submit

**Expected Result**:
- Question created successfully
- Question response includes:
  ```json
  {
    "meta": {
      "image_url": "/api/v1/files/serve/{file-id}",
      "image_file_id": "{file-id}",
      ...
    }
  }
  ```

### Test 2: Add question with image via API (curl)

```bash
curl -X POST "http://localhost:8080/api/v1/quizzes/{quiz-id}/questions" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "image=@test-image.png" \
  -F 'question_data={
    "question_text": "Test question with image",
    "question_type": "multiple_choice",
    "points": 1,
    "answers": [
      {"answer_text": "Option A", "is_correct": true},
      {"answer_text": "Option B", "is_correct": false}
    ]
  }'
```

**Expected Response**:
```json
{
  "status": "success",
  "data": {
    "id": "...",
    "question_text": "Test question with image",
    "meta": {
      "image_url": "/api/v1/files/serve/abc-123",
      "image_file_id": "abc-123",
      "answer_count": 2
    }
  }
}
```

### Test 3: View uploaded image

```bash
curl -X GET "http://localhost:8080/api/v1/files/serve/{file-id}" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  --output downloaded-image.png
```

**Expected Result**:
- Image file downloads successfully
- Can be opened and viewed

### Test 4: Update question image

```bash
curl -X PATCH "http://localhost:8080/api/v1/questions/{question-id}" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "image=@new-image.jpg" \
  -F 'question_data={
    "question_text": "Updated question text"
  }'
```

**Expected Result**:
- Question updated successfully
- Meta now shows new image_file_id

## How It Works (Flow)

### Frontend (already implemented)
1. User fills out question form and selects image file
2. `QuestionFormSingle.jsx` captures image as `File` object
3. `QuestionServices.js` `addQuestion()` detects image file
4. Creates `FormData` with:
   - `image` field: the actual File object
   - `question_data` field: JSON string with all question data
5. Sends as `multipart/form-data` request

### Backend (now fixed)
1. `AddQuestion` handler receives multipart request
2. Extracts and saves image file
3. Stores the uploaded file ID in a variable
4. Parses `question_data` JSON into request struct
5. **NEW**: Restores the uploaded file ID (prevents it from being lost)
6. Passes to core layer with correct `ImageFileID`
7. Core layer (`quiz/core/quiz.go`):
   - Validates file exists and user owns it
   - Adds to meta: `image_url` and `image_file_id`
8. Returns question with complete metadata

## Benefits

✅ **Single-step upload**: Frontend uploads image directly with question, no pre-upload needed
✅ **Automatic metadata**: Backend automatically populates image_url and image_file_id
✅ **File validation**: Backend verifies file exists and user has permission
✅ **Consistent behavior**: Works same for create and update operations
✅ **Backward compatible**: Still supports JSON-only requests with pre-uploaded image_file_id

## Notes

- The `AddQuestionsBulk` handler did NOT have this bug because it parses JSON first, then uploads images after
- Frontend code was already correct and didn't need any changes
- This fix completes the single-step image upload feature for manual question creation
