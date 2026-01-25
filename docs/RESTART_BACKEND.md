# Backend Restart Required

## What Was Fixed

The image upload bug in manual question creation has been fixed. The issue was that when processing multipart form data, the backend was not properly preserving the uploaded image file ID in BOTH code paths:

1. ✅ **question_data JSON path** - Fixed
2. ✅ **Individual form fields path** - Fixed

## Files Modified

- `backend/api/quiz.go` - AddQuestion handler (lines 455-535)
- `backend/api/quiz.go` - UpdateQuestion handler (lines 746-824)
- `backend/quiz/core/quiz.go` - Added debug logging (line 455)

## How to Restart

### Stop Current Backend

```bash
# Find the process
ps aux | grep "go run main.go"

# Or on Windows
tasklist | findstr "go"

# Kill it (Linux/Mac)
pkill -f "go run main.go"

# Or on Windows
taskkill /F /IM go.exe
```

### Start Fresh Backend

```bash
cd backend
go run main.go
```

## Test the Fix

After restarting, try creating a question with an image:

1. Go to your quiz edit page
2. Click "Add Questions Manually"
3. Select any question type
4. Fill in the question details
5. **Upload an image**
6. Submit

**Expected Response:**
```json
{
  "status": "success",
  "message": "Question added successfully",
  "data": {
    "id": "...",
    "question_text": "your question",
    "meta": {
      "image_url": "/api/v1/files/serve/abc-123",
      "image_file_id": "abc-123",
      "answer_count": 2
    }
  }
}
```

## Logs to Watch

With the new logging, you should see:

```
[api.AddQuestion [user_id=... [quiz_id=... [has_image=true [image_file_id=abc-123
[core.AddQuestion::ProcessingImage [image_file_id=abc-123
[core.AddQuestion [attached_image=abc-123
```

If you DON'T see these logs when uploading an image, there's still an issue.

## What the Frontend Sends

The frontend (`QuestionServices.js`) already sends the correct format:

```javascript
const formData = new FormData();
formData.append('image', questionData.image_file);  // The actual File object
formData.append('question_data', JSON.stringify({
  question_text: "...",
  question_type: "...",
  points: 1,
  answers: [...]
}));
```

## What the Backend Does Now

1. **Extract image file** from form → Upload to disk → Get file ID
2. **Store file ID** in `uploadedImageFileID` variable
3. **Parse question_data** JSON into `req` struct
4. **Restore image file ID** to `req.ImageFileID` (THIS WAS MISSING BEFORE)
5. Pass to core layer with correct ImageFileID
6. Core layer adds `image_url` and `image_file_id` to meta

## Common Issues

### Image still not in meta

Check logs to see:
- Is the image being uploaded? (Look for "ImageUploadErr")
- Is the ImageFileID being set? (Look for "has_image=true")
- Is core layer processing it? (Look for "ProcessingImage")

### "image-file-not-found" error

The uploaded file doesn't exist in database. Check:
- File was successfully saved
- File ID is valid UUID
- User owns the file

### "not-image-file-owner" error

The file exists but belongs to another user.
