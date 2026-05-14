import React, { useState } from "react"
import { AddCommentDocument } from "../gql/graphql"
import { useMutation } from 'urql'

type CommentFormProps = {
    requestID: string;
    onSuccess: () => void;
    onDemoSubmit?: (content: string) => void;
};

export function CommentForm({ requestID, onSuccess, onDemoSubmit }: CommentFormProps) {
    const [content, setContent] = useState("")
    const [error, setError] = useState<string | null>(null)
    const [{ fetching: addCommentFetching }, addComment] = useMutation(
        AddCommentDocument,
    )
    
    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setError(null);
        
        if (!content.trim()) {
            setError('コメントを入力してください');
            return;
        }
        
        if (onDemoSubmit) {
            // デモモード
            onDemoSubmit(content);
            setContent('');
            onSuccess();
        } else {
            // プロダクションモード
            const result = await addComment({ requestID, content });
            if (result.error) {
                setError(result.error.message);
            } else {
                setContent('');
                onSuccess();
            }
        }
    };

    return (
        <div className="commentFormContainer">
            {error && <div className="errorMessage">{error}</div>}
            <form className="commentForm" onSubmit={handleSubmit}>
                <textarea 
                    name="content"
                    placeholder="コメントを入力してください"
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    rows={3}
                    disabled={addCommentFetching}
                /> 
                <button 
                    type="submit" 
                    disabled={addCommentFetching || !content.trim()}
                    className="submitButton"
                >
                    {addCommentFetching ? '送信中...' : 'コメントを送信'}
                </button>
            </form>
        </div>
    )
}