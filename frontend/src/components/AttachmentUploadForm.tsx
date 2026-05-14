import { useState } from 'react'
import { useMutation } from 'urql'
import { UploadAttachmentDocument } from '../gql/graphql'

type AttachmentUploadFormProps = {
    requestID: string;
    onSuccess: () => void;
};

export default function AttachmentUploadForm({ requestID, onSuccess }: AttachmentUploadFormProps) {
    const [file, setFile] = useState<File | null>(null)
    const [{ fetching }, uploadAttachment] = useMutation(UploadAttachmentDocument)

    const handleSubmit = async(e: React.FormEvent) => {
        e.preventDefault();
        if (!file) {
            alert('ファイルを選択してください');
            return
        }

        const result = await uploadAttachment({
             requestID,
             file,
        })

        if (result.error) {
            alert(`アップロードエラー: ${result.error.message}`);
            return
        }

        setFile(null)

        const input = document.getElementById('file-input') as HTMLInputElement
        if (input) input.value = ''

        if (onSuccess) onSuccess()
    }

    return (
        <form onSubmit={handleSubmit} style={{ marginTop: 10}}>
            <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
                <input
                    id="file-input"
                    type="file"
                    onChange={(e) => setFile(e.target.files?.[0] || null)}
                    disabled={fetching}
                />
                <button
                    type="submit"
                    className="btn"
                    disabled={fetching || !file}
                >
                    {fetching ? 'アップロード中...' : 'アップロード'}
                </button>
            </div>
            {file && (
                <p className="note" style={{ marginTop:5 }}>
                    選択：{file.name} ({(file.size / 1024).toFixed(1)} KB)
                </p>
            )}
        </form>
    )
}

