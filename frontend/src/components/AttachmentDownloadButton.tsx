import { useEffect, useState } from 'react'
import { useQuery } from 'urql'
import { AttachmentDownloadUrlDocument } from '../gql/graphql'

type AttachmentDownloadButtonProps = {
    requestID: string;
    attachmentID: string;
    filename: string;
};

export default function AttachmentDownloadButton({ requestID, attachmentID, filename }: AttachmentDownloadButtonProps) {
    const [shouldFetch, setShouldFetch] = useState(false);
    
    const [{ data, fetching, error }] = useQuery({
        query: AttachmentDownloadUrlDocument,
        variables: { requestID, attachmentID },
        pause: !shouldFetch,
    });

    useEffect(() => {
        if (data?.attachmentDownloadURL && shouldFetch) {
            window.open(data.attachmentDownloadURL, '_blank');
            setShouldFetch(false);
        } else if (error && shouldFetch) {
            alert(`ダウンロードエラー: ${error.message}`);
            setShouldFetch(false);
        }
    }, [data, error, shouldFetch]);

    const handleClick = () => {
        setShouldFetch(true);
    };

    return (
        <button
            className="btn btn--ghost"
            onClick={handleClick}
            disabled={fetching}
        >
            {fetching ? 'ダウンロード中...' : 'ダウンロード'}
        </button>
    );
}