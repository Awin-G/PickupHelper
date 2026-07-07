import { View, Text, Image } from '@tarojs/components';
import { Button } from '@nutui/nutui-react-taro';
import './index.scss';

interface EmptyStateProps {
  icon?: string;
  title: string;
  description?: string;
  buttonText?: string;
  onButtonClick?: () => void;
}

export default function EmptyState({
  icon,
  title,
  description,
  buttonText,
  onButtonClick,
}: EmptyStateProps) {
  return (
    <View className='empty-state'>
      {icon && <Image className='empty-state__icon' src={icon} mode='aspectFit' />}
      <Text className='empty-state__title'>{title}</Text>
      {description && <Text className='empty-state__desc'>{description}</Text>}
      {buttonText && onButtonClick && (
        <Button type='primary' size='small' className='empty-state__btn' onClick={onButtonClick}>
          {buttonText}
        </Button>
      )}
    </View>
  );
}
