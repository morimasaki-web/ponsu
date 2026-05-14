import { RequestCommentsDocument } from "../gql/graphql";
import { useQuery } from 'urql'
import { DemoComment } from '../demoData'

type CommentListProps = {
    requestID: string;
    demoComments?: DemoComment[];
};

export function CommentList({ requestID, demoComments }: CommentListProps) {
    const [{ data, fetching, error }] = useQuery({
        query: RequestCommentsDocument,
        variables: { requestID },
        requestPolicy: 'cache-and-network',
        pause: !!demoComments, // デモモードの場合はGraphQLクエリを停止
    });

    const comments = demoComments ?? data?.comments;

    return (
        <div className="commentList">
            {!demoComments && fetching && <p className="loading">読み込み中...</p>}
            {!demoComments && error && <p className="error">コメントの読み込み中にエラーが発生しました: {error.message}</p>}
            {comments && comments.length === 0 && (
                <p className="noComments">まだコメントがありません</p>
            )}
            {comments && comments.length > 0 && (
                <ul className="comments">    
                    {comments.map((comment) => (
                        <li key={comment.id} className="comment">
                            <div className="commentHeader">
                                <span className="commentAuthor">{comment.userID}</span>
                                <span className="commentDate">{new Date(comment.createdAt).toLocaleString()}</span>
                            </div>
                            <div className="commentContent">{comment.content}</div>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}